package impl

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/model"
	authdomain "github.com/yourname/pos-koperasi/internal/modules/auth/domain"
	productdomain "github.com/yourname/pos-koperasi/internal/modules/product/domain"
	"github.com/yourname/pos-koperasi/internal/modules/chat/contract"
	"github.com/yourname/pos-koperasi/internal/modules/chat/utility"
)

type Service struct {
	db   *sqlx.DB
	repo contract.ChatRepository
	hub  *utility.Hub
	pres *utility.Presence
}

func New(db *sqlx.DB, repo contract.ChatRepository, hub *utility.Hub, pres *utility.Presence) *Service {
	return &Service{db: db, repo: repo, hub: hub, pres: pres}
}

var _ contract.ChatService = (*Service)(nil)

func (s *Service) CreateConversation(memberID int, input contract.CreateConversationInput) (*model.ChatConversation, error) {
	if input.ProductID <= 0 || input.SellerEmployeeID <= 0 {
		return nil, fmt.Errorf("product_id dan seller_employee_id wajib")
	}
	first := strings.TrimSpace(input.FirstMessage)
	if first == "" {
		return nil, fmt.Errorf("first_message wajib — percakapan disimpan setelah Anda mengirim pesan pertama")
	}

	var emp authdomain.Employee
	if err := s.db.Get(&emp, `SELECT * FROM employees WHERE id = $1 AND is_active = true`, input.SellerEmployeeID); err != nil {
		return nil, fmt.Errorf("penjual (karyawan) tidak ditemukan atau nonaktif")
	}

	var p productdomain.Product
	if err := s.db.Get(&p, `SELECT * FROM products WHERE id = $1 AND is_active = true`, input.ProductID); err != nil {
		return nil, fmt.Errorf("produk tidak ditemukan atau nonaktif")
	}
	_ = p

	bid := emp.BranchID
	pid := input.ProductID
	conv := model.ChatConversation{
		ProductID:        &pid,
		MemberID:         memberID,
		SellerEmployeeID: input.SellerEmployeeID,
		BranchID:         &bid,
	}

	id, err := s.repo.CreateConversation(conv)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat percakapan: %w", err)
	}

	// Lampiran produk + pesan pertama dalam satu alur (baru tersimpan di DB sekarang).
	ins, err := s.repo.ShouldInsertProductContext(id, input.ProductID)
	if err == nil && ins {
		pcopy := input.ProductID
		saved, err := s.repo.InsertMessage(id, "buyer", "", &pcopy)
		if err != nil {
			return nil, err
		}
		out, _ := json.Marshal(map[string]any{
			"type":              "message",
			"conversation_id": id,
			"message":           saved,
		})
		s.hub.BroadcastRoom(id, out)
	}

	savedText, err := s.repo.InsertMessage(id, "buyer", first, nil)
	if err != nil {
		return nil, err
	}
	outText, _ := json.Marshal(map[string]any{
		"type":              "message",
		"conversation_id": id,
		"message":           savedText,
	})
	s.hub.BroadcastRoom(id, outText)

	out, err := s.repo.GetConversation(id)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) ListConversations(p *middleware.ChatPrincipal) ([]model.ChatConversation, error) {
	if p.IsEmployee {
		return s.repo.ListForEmployee(p.EmployeeID)
	}
	return s.repo.ListForMember(p.MemberID)
}

func (s *Service) ListMessages(convID int64, p *middleware.ChatPrincipal) ([]model.ChatMessage, error) {
	if err := s.assertParticipant(convID, p); err != nil {
		return nil, err
	}
	return s.repo.ListMessages(convID, 100)
}

func (s *Service) assertParticipant(convID int64, p *middleware.ChatPrincipal) error {
	var mid *int
	var eid *int
	if !p.IsEmployee && p.MemberID != 0 {
		v := p.MemberID
		mid = &v
	}
	if p.IsEmployee {
		v := p.EmployeeID
		eid = &v
	}
	_, err := s.repo.ParticipantRole(convID, mid, eid)
	return err
}

// JoinRoom memverifikasi peserta lalu mendaftarkan klien WS ke room.
func (s *Service) JoinRoom(convID int64, p *middleware.ChatPrincipal, c *utility.Client) error {
	if err := s.assertParticipant(convID, p); err != nil {
		return err
	}
	s.hub.JoinRoom(convID, c)
	return nil
}

// PostMessage menyimpan pesan dan mengembalikan payload broadcast JSON.
func (s *Service) PostMessage(convID int64, body string, p *middleware.ChatPrincipal) ([]byte, error) {
	if body == "" {
		return nil, fmt.Errorf("pesan kosong")
	}

	var mid *int
	var eid *int
	if !p.IsEmployee && p.MemberID != 0 {
		v := p.MemberID
		mid = &v
	}
	if p.IsEmployee {
		v := p.EmployeeID
		eid = &v
	}

	role, err := s.repo.ParticipantRole(convID, mid, eid)
	if err != nil {
		return nil, err
	}

	saved, err := s.repo.InsertMessage(convID, role, body, nil)
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"type":            "message",
		"conversation_id": convID,
		"message":         saved,
	}
	return json.Marshal(out)
}

func (s *Service) Hub() *utility.Hub {
	return s.hub
}

func (s *Service) PresenceConnect(p *middleware.ChatPrincipal) {
	if s.pres == nil || p == nil {
		return
	}
	if p.IsEmployee {
		s.pres.ConnectEmployee(p.EmployeeID)
		return
	}
	s.pres.ConnectMember(p.MemberID)
}

func (s *Service) PresenceDisconnect(p *middleware.ChatPrincipal) {
	if s.pres == nil || p == nil {
		return
	}
	if p.IsEmployee {
		s.pres.DisconnectEmployee(p.EmployeeID)
		return
	}
	s.pres.DisconnectMember(p.MemberID)
}

func (s *Service) ConversationPresence(convID int64, p *middleware.ChatPrincipal) (contract.ConversationPresence, error) {
	var out contract.ConversationPresence
	if err := s.assertParticipant(convID, p); err != nil {
		return out, err
	}
	c, err := s.repo.GetConversation(convID)
	if err != nil {
		return out, err
	}
	if s.pres != nil {
		out.BuyerOnline = s.pres.IsMemberOnline(c.MemberID)
		out.SellerOnline = s.pres.IsEmployeeOnline(c.SellerEmployeeID)
	}
	return out, nil
}

