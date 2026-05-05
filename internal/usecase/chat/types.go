package chat

import (
	"github.com/fathanazka354/pos-koperasi/internal/entity/auth"
	"github.com/fathanazka354/pos-koperasi/internal/entity/product"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/chatws/utility"
	"github.com/fathanazka354/pos-koperasi/internal/model"
)

// ChatPrincipal identitas untuk HTTP/WebSocket chat (karyawan atau member).
type ChatPrincipal struct {
	IsEmployee bool
	EmployeeID int
	BranchID   int
	Role       string
	MemberID   int
}

type CreateConversationInput struct {
	ProductID        int
	SellerEmployeeID int
	FirstMessage     string // wajib — percakapan + lampiran produk baru tersimpan setelah pesan pertama
}

// ConversationPresence status online lawan bicara (ada koneksi WS chat).
type ConversationPresence struct {
	BuyerOnline  bool
	SellerOnline bool
}

type Repository interface {
	CreateConversation(c model.ChatConversation) (int64, error)
	GetConversation(id int64) (*model.ChatConversation, error)
	ListForMember(memberID int) ([]model.ChatConversation, error)
	ListForEmployee(employeeID int) ([]model.ChatConversation, error)
	InsertMessage(convID int64, senderRole, body string, productID *int) (*model.ChatMessage, error)
	ShouldInsertProductContext(convID int64, productID int) (bool, error)
	ListMessages(convID int64, limit int) ([]model.ChatMessage, error)
	ParticipantRole(convID int64, memberID *int, employeeID *int) (string, error)
	GetActiveEmployeeByID(id int) (*auth.Employee, error)
	GetActiveProductByID(id int) (*product.Product, error)
}

type Usecase interface {
	CreateConversation(memberID int, input CreateConversationInput) (*model.ChatConversation, error)
	ListConversations(p *ChatPrincipal) ([]model.ChatConversation, error)
	ListMessages(convID int64, p *ChatPrincipal) ([]model.ChatMessage, error)
	JoinRoom(convID int64, p *ChatPrincipal, c *utility.Client) error
	PostMessage(convID int64, body string, p *ChatPrincipal) ([]byte, error)
	ConversationPresence(convID int64, p *ChatPrincipal) (ConversationPresence, error)
	PresenceConnect(p *ChatPrincipal)
	PresenceDisconnect(p *ChatPrincipal)
	Hub() *utility.Hub
}

