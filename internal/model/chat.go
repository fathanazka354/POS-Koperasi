package model

import "time"

// ChatConversation satu thread per pasangan pembeli–penjual; product_id = konteks terakhir (opsional).
type ChatConversation struct {
	ID                 int64      `db:"id"                    json:"id"`
	ProductID          *int       `db:"product_id"            json:"product_id,omitempty"`
	MemberID           int        `db:"member_id"             json:"member_id"`
	SellerEmployeeID   int        `db:"seller_employee_id"    json:"seller_employee_id"`
	BranchID           *int       `db:"branch_id"             json:"branch_id,omitempty"`
	LastMessageAt      *time.Time `db:"last_message_at"       json:"last_message_at,omitempty"`
	CreatedAt          time.Time  `db:"created_at"            json:"created_at"`
	ProductName        string     `db:"product_name"          json:"product_name,omitempty"`
	MemberName         string     `db:"member_name"           json:"member_name,omitempty"`
	SellerName         string     `db:"seller_name"           json:"seller_name,omitempty"`
	LastMessage        *string    `db:"last_message"          json:"last_message,omitempty"` // subquery bisa NULL bila belum ada pesan
}

type ChatMessage struct {
	ID             int64     `db:"id"              json:"id"`
	ConversationID int64     `db:"conversation_id" json:"conversation_id"`
	SenderRole     string    `db:"sender_role"     json:"sender_role"`
	Body           string    `db:"body"            json:"body"`
	ProductID      *int      `db:"product_id"      json:"product_id,omitempty"`
	CreatedAt      time.Time `db:"created_at"      json:"created_at"`
}
