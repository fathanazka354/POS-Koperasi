package contract

import (
	"github.com/yourname/pos-koperasi/internal/midtrans"
	authdomain "github.com/yourname/pos-koperasi/internal/modules/auth/domain"
	productdomain "github.com/yourname/pos-koperasi/internal/modules/product/domain"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/domain"
)

type TransactionItemInput struct {
	ProductID int
	Quantity  int
}

type CreateTransactionInput struct {
	BranchID   int
	RegisterID int
	MemberCode string
	Items      []TransactionItemInput
	PromoCode  string
	PayMethod  string // cash | qris | va | ewallet
}

type MidtransAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

type CreateTransactionOutput struct {
	Transaction domain.Transaction `json:"transaction"`
	PaymentURL  string            `json:"payment_url,omitempty"`
	QRString    string            `json:"qr_string,omitempty"`
	VANumber    string            `json:"va_number,omitempty"`

	MidtransTransactionID     string          `json:"midtrans_transaction_id,omitempty"`
	MidtransTransactionStatus string          `json:"midtrans_transaction_status,omitempty"`
	MidtransPaymentType       string          `json:"midtrans_payment_type,omitempty"`
	MidtransActions           []MidtransAction `json:"midtrans_actions,omitempty"`
	MidtransRawResponse       string          `json:"midtrans_raw_response,omitempty"`

	Message string `json:"message"`
}

type ProcessTransactionInput struct {
	Transaction domain.Transaction
	Items       []domain.TransactionItem
	BranchID    int
}

type SettleInput struct {
	TransactionID int64
	BranchID      int
	MemberID      *int
	GrandTotal    float64
	Payment       domain.Payment
}

type TransactionRepository interface {
	GetProductWithStock(productID, branchID int) (*productdomain.ProductWithStock, error)
	GetMemberByCode(memberCode string) (*authdomain.Member, error)
	CreatePendingTransaction(input ProcessTransactionInput) (int64, error)
	GetTransactionByID(txID int64) (*domain.Transaction, error)
	GetItemsByTransactionID(txID int64) ([]domain.TransactionItem, error)
	GetTransactionsByMember(memberID int) ([]domain.Transaction, error)
	SettleTransaction(input SettleInput) error
	CancelTransaction(transactionID int64) error
	PaymentExistsByRef(referenceNo string) (bool, error)
	GetTransactionByInvoice(invoiceNo string) (*domain.Transaction, error)
}

type TransactionService interface {
	CreateTransaction(input CreateTransactionInput, employeeID int) (*CreateTransactionOutput, error)
	HandleMidtransNotification(payload midtrans.NotificationPayload) error
}

