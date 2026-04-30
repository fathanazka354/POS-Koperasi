package dto

type TransactionItemInput struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type CreateTransactionRequest struct {
	BranchID   int                  `json:"branch_id"`
	RegisterID int                  `json:"register_id"`
	MemberCode string               `json:"member_code"`
	Items      []TransactionItemInput `json:"items"`
	PromoCode  string               `json:"promo_code"`
	PayMethod  string               `json:"pay_method"` // cash | qris | va | ewallet
}

