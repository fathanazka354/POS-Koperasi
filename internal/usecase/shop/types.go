package shop

import (
	txcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/transaction"
	txdomain "github.com/fathanazka354/pos-koperasi/internal/entity/transaction"
	voucherdomain "github.com/fathanazka354/pos-koperasi/internal/entity/voucher"
)

type CheckoutItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type CheckoutInput struct {
	MemberID    int            `json:"-"`
	AddressID   int            `json:"address_id"`
	VoucherCode string         `json:"voucher_code"`
	Items       []CheckoutItem `json:"items"`
	PayMethod   string         `json:"pay_method"` // cash | qris | va | ewallet
}

type CheckoutOutput struct {
	*txcontract.CreateTransactionOutput
	VoucherDiscount float64 `json:"voucher_discount"`
}

type TransactionDetail struct {
	Transaction txdomain.Transaction       `json:"transaction"`
	Items       []txdomain.TransactionItem `json:"items"`
}

// VoucherValidateResult dikembalikan dari endpoint validasi voucher.
type VoucherValidateResult struct {
	Voucher     *voucherdomain.Voucher `json:"voucher"`
	DiscountAmt float64                `json:"discount_amount"`
	Label       string                 `json:"label"`
}

type Usecase interface {
	Checkout(input CheckoutInput) (*CheckoutOutput, error)
	GetTransactionsByMember(memberID int) ([]txdomain.Transaction, error)
	GetTransactionDetail(invoiceNo string, memberID int) (*TransactionDetail, error)
	CheckPaymentStatus(invoiceNo string, memberID int) (string, error)

	// Voucher
	ValidateVoucher(code string, subtotal float64) (*VoucherValidateResult, error)
	ListVouchers() ([]*voucherdomain.Voucher, error)
}

