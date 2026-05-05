package transaction

import "time"

type Transaction struct {
	ID              int64     `db:"id"               json:"id"`
	BranchID        int       `db:"branch_id"        json:"branch_id"`
	RegisterID      int       `db:"register_id"      json:"register_id"`
	EmployeeID      *int      `db:"employee_id"      json:"employee_id"`
	MemberID        *int      `db:"member_id"        json:"member_id"`
	PromoID         *int      `db:"promo_id"         json:"promo_id"`
	AddressID       *int      `db:"address_id"       json:"address_id,omitempty"`
	VoucherID       *int      `db:"voucher_id"       json:"voucher_id,omitempty"`
	VoucherDiscount float64   `db:"voucher_discount" json:"voucher_discount"`
	InvoiceNo       string    `db:"invoice_no"       json:"invoice_no"`
	Subtotal        float64   `db:"subtotal"         json:"subtotal"`
	Discount        float64   `db:"discount"         json:"discount"`
	Tax             float64   `db:"tax"              json:"tax"`
	GrandTotal      float64   `db:"grand_total"      json:"grand_total"`
	Status          string    `db:"status"           json:"status"` // pending | paid | cancelled | refunded
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"       json:"updated_at,omitempty"`
}

type TransactionItem struct {
	ID            int     `db:"id"             json:"id"`
	TransactionID int64   `db:"transaction_id" json:"transaction_id"`
	ProductID     int     `db:"product_id"     json:"product_id"`
	Quantity      int     `db:"quantity"       json:"quantity"`
	UnitPrice     float64 `db:"unit_price"     json:"unit_price"`
	Discount      float64 `db:"discount"       json:"discount"`
	Subtotal      float64 `db:"subtotal"       json:"subtotal"`
}

type Payment struct {
	ID            int        `db:"id"             json:"id"`
	TransactionID int64      `db:"transaction_id" json:"transaction_id"`
	Method        string     `db:"method"         json:"method"` // cash | qris | va | ewallet
	Amount        float64    `db:"amount"         json:"amount"`
	ChangeAmount  float64    `db:"change_amount"  json:"change_amount"`
	ReferenceNo   string     `db:"reference_no"   json:"reference_no"`
	PaidAt        *time.Time `db:"paid_at"        json:"paid_at"`
}

