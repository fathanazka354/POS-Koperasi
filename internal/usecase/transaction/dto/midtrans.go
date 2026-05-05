package dto

// MidtransNotification payload bisnis untuk penyelesaian pembayaran (DTO agar tidak membuat import cycle).
type MidtransNotification struct {
	OrderID           string `json:"order_id"`
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	PaymentType       string `json:"payment_type"`
	GrossAmount       string `json:"gross_amount"`
}

