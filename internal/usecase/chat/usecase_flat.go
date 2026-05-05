package chat

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fathanazka354/pos-koperasi/internal/gateway/chatws/utility"
	"github.com/fathanazka354/pos-koperasi/internal/model"
)

type usecase struct {
	repo Repository
	hub  *utility.Hub
	pres *utility.Presence
}

func New(repo Repository, hub *utility.Hub, pres *utility.Presence) Usecase {
	return &usecase{repo: repo, hub: hub, pres: pres}
}

func (u *usecase) CreateConversation(memberID int, input CreateConversationInput) (*model.ChatConversation, error) {
	if input.ProductID <= 0 || input.SellerEmployeeID <= 0 {
		return nil, fmt.Errorf("product_id dan seller_employee_id wajib")
	}
	first := strings.TrimSpace(input.FirstMessage)
	if first == "" {
		return nil, fmt.Errorf("first_message wajib — percakapan disimpan setelah Anda mengirim pesan pertama")
	}

	emp, err := u.repo.GetActiveEmployeeByID(input.SellerEmployeeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("penjual (karyawan) tidak ditemukan atau nonaktif")
		}
		return nil, err
	}

	_, err = u.repo.GetActiveProductByID(input.ProductID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("produk tidak ditemukan atau nonaktif")
		}
		return nil, err
	}

	bid := emp.BranchID
	pid := input.ProductID
	conv := model.ChatConversation{
		ProductID:        &pid,
		MemberID:         memberID,
		SellerEmployeeID: input.SellerEmployeeID,
		BranchID:         &bid,
	}

	id, err := u.repo.CreateConversation(conv)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat percakapan: %w", err)
	}

	ins, err := u.repo.ShouldInsertProductContext(id, input.ProductID)
	if err == nil && ins {
		pcopy := input.ProductID
		saved, err := u.repo.InsertMessage(id, "buyer", "", &pcopy)
		if err != nil {
			return nil, err
		}
		out, _ := json.Marshal(map[string]any{
			"type":            "message",
			"conversation_id": id,
			"message":         saved,
		})
		u.hub.BroadcastRoom(id, out)
	}

	savedText, err := u.repo.InsertMessage(id, "buyer", first, nil)
	if err != nil {
		return nil, err
	}
	outText, _ := json.Marshal(map[string]any{
		"type":            "message",
		"conversation_id": id,
		"message":         savedText,
	})
	u.hub.BroadcastRoom(id, outText)

	out, err := u.repo.GetConversation(id)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *usecase) ListConversations(p *ChatPrincipal) ([]model.ChatConversation, error) {
	if p.IsEmployee {
		return u.repo.ListForEmployee(p.EmployeeID)
	}
	return u.repo.ListForMember(p.MemberID)
}

func (u *usecase) ListMessages(convID int64, p *ChatPrincipal) ([]model.ChatMessage, error) {
	if err := u.assertParticipant(convID, p); err != nil {
		return nil, err
	}
	return u.repo.ListMessages(convID, 100)
}

func (u *usecase) assertParticipant(convID int64, p *ChatPrincipal) error {
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
	_, err := u.repo.ParticipantRole(convID, mid, eid)
	return err
}

// JoinRoom memverifikasi peserta lalu mendaftarkan klien WS ke room.
func (u *usecase) JoinRoom(convID int64, p *ChatPrincipal, c *utility.Client) error {
	if err := u.assertParticipant(convID, p); err != nil {
		return err
	}
	u.hub.JoinRoom(convID, c)
	return nil
}

// PostMessage menyimpan pesan dan mengembalikan payload broadcast JSON.
func (u *usecase) PostMessage(convID int64, body string, p *ChatPrincipal) ([]byte, error) {
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

	role, err := u.repo.ParticipantRole(convID, mid, eid)
	if err != nil {
		return nil, err
	}

	saved, err := u.repo.InsertMessage(convID, role, body, nil)
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

func (u *usecase) Hub() *utility.Hub { return u.hub }

func (u *usecase) PresenceConnect(p *ChatPrincipal) {
	if u.pres == nil || p == nil {
		return
	}
	if p.IsEmployee {
		u.pres.ConnectEmployee(p.EmployeeID)
		return
	}
	u.pres.ConnectMember(p.MemberID)
}

func (u *usecase) PresenceDisconnect(p *ChatPrincipal) {
	if u.pres == nil || p == nil {
		return
	}
	if p.IsEmployee {
		u.pres.DisconnectEmployee(p.EmployeeID)
		return
	}
	u.pres.DisconnectMember(p.MemberID)
}

func (u *usecase) ConversationPresence(convID int64, p *ChatPrincipal) (ConversationPresence, error) {
	var out ConversationPresence
	if err := u.assertParticipant(convID, p); err != nil {
		return out, err
	}
	c, err := u.repo.GetConversation(convID)
	if err != nil {
		return out, err
	}
	if u.pres != nil {
		out.BuyerOnline = u.pres.IsMemberOnline(c.MemberID)
		out.SellerOnline = u.pres.IsEmployeeOnline(c.SellerEmployeeID)
	}
	return out, nil
}

