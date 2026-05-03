package contract

import (
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/model"
	"github.com/yourname/pos-koperasi/internal/modules/chat/utility"
)

type CreateConversationInput struct {
	ProductID        int
	SellerEmployeeID int
	FirstMessage     string // wajib — percakapan + lampiran produk baru tersimpan setelah pesan pertama
}

type ChatRepository interface {
	CreateConversation(c model.ChatConversation) (int64, error)
	GetConversation(id int64) (*model.ChatConversation, error)
	ListForMember(memberID int) ([]model.ChatConversation, error)
	ListForEmployee(employeeID int) ([]model.ChatConversation, error)
	InsertMessage(convID int64, senderRole, body string, productID *int) (*model.ChatMessage, error)
	// ShouldInsertProductContext false jika baris terakhir sudah stub produk yang sama (hindari duplikat beruntun).
	ShouldInsertProductContext(convID int64, productID int) (bool, error)
	ListMessages(convID int64, limit int) ([]model.ChatMessage, error)
	ParticipantRole(convID int64, memberID *int, employeeID *int) (string, error)
}

// ConversationPresence status online lawan bicara (ada koneksi WS chat).
type ConversationPresence struct {
	BuyerOnline  bool
	SellerOnline bool
}

type ChatService interface {
	CreateConversation(memberID int, input CreateConversationInput) (*model.ChatConversation, error)
	ListConversations(p *middleware.ChatPrincipal) ([]model.ChatConversation, error)
	ListMessages(convID int64, p *middleware.ChatPrincipal) ([]model.ChatMessage, error)
	JoinRoom(convID int64, p *middleware.ChatPrincipal, c *utility.Client) error
	PostMessage(convID int64, body string, p *middleware.ChatPrincipal) ([]byte, error)
	ConversationPresence(convID int64, p *middleware.ChatPrincipal) (ConversationPresence, error)
	PresenceConnect(p *middleware.ChatPrincipal)
	PresenceDisconnect(p *middleware.ChatPrincipal)
	Hub() *utility.Hub
}

