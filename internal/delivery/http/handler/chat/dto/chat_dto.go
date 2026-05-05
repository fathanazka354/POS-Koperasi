package dto

import "github.com/fathanazka354/pos-koperasi/internal/model"

type CreateConversationRequest struct {
	ProductID        int    `json:"product_id"`
	SellerEmployeeID int    `json:"seller_employee_id"`
	FirstMessage     string `json:"first_message"`
}

type WSClientMessage struct {
	Type           string `json:"type"`
	ConversationID int64  `json:"conversation_id"`
	Body           string `json:"body"`
}

type WSServerMessage struct {
	Type           string             `json:"type"`
	ConversationID int64              `json:"conversation_id,omitempty"`
	Message        *model.ChatMessage `json:"message,omitempty"`
	Error          string             `json:"error,omitempty"`
}

