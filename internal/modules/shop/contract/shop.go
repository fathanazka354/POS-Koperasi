package contract

import (
	"github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	txdomain "github.com/yourname/pos-koperasi/internal/modules/transaction/domain"
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
	Transaction txdomain.Transaction   `json:"transaction"`
	Items       []txdomain.TransactionItem `json:"items"`
}

type ShopService interface {
	Checkout(input CheckoutInput) (*CheckoutOutput, error)
	GetTransactionsByMember(memberID int) ([]txdomain.Transaction, error)
	GetTransactionDetail(invoiceNo string, memberID int) (*TransactionDetail, error)
	CheckPaymentStatus(invoiceNo string, memberID int) (string, error)
}
