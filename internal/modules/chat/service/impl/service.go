package impl

import (
	"encoding/json"
	"fmt"

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
}

func New(db *sqlx.DB, repo contract.ChatRepository, hub *utility.Hub) *Service {
	return &Service{db: db, repo: repo, hub: hub}
}

var _ contract.ChatService = (*Service)(nil)

func (s *Service) CreateConversation(memberID int, input contract.CreateConversationInput) (*model.ChatConversation, error) {
	if input.ProductID <= 0 || input.SellerEmployeeID <= 0 {
		return nil, fmt.Errorf("product_id dan seller_employee_id wajib")
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
	conv := model.ChatConversation{
		ProductID:        input.ProductID,
		MemberID:         memberID,
		SellerEmployeeID: input.SellerEmployeeID,
		BranchID:         &bid,
	}

	id, err := s.repo.CreateConversation(conv)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat percakapan: %w", err)
	}

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

	saved, err := s.repo.InsertMessage(convID, role, body)
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

