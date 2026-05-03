package contract

import (
	"github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	txdomain "github.com/yourname/pos-koperasi/internal/modules/transaction/domain"
	voucherdomain "github.com/yourname/pos-koperasi/internal/modules/voucher/domain"
)

type CheckoutItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type CheckoutInput struct {
	MemberID    int
	AddressID   int            `json:"address_id"`
	VoucherCode string         `json:"voucher_code"`
	Items       []CheckoutItem `json:"items"`
	PayMethod   string         `json:"pay_method"` // cash | qris | va | ewallet
}

type CheckoutOutput struct {
	*contract.CreateTransactionOutput
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

type ShopService interface {
	Checkout(input CheckoutInput) (*CheckoutOutput, error)
	GetTransactionsByMember(memberID int) ([]txdomain.Transaction, error)
	GetTransactionDetail(invoiceNo string, memberID int) (*TransactionDetail, error)
	CheckPaymentStatus(invoiceNo string, memberID int) (string, error)
	// Voucher
	ValidateVoucher(code string, subtotal float64) (*VoucherValidateResult, error)
	ListVouchers() ([]*voucherdomain.Voucher, error)
}
