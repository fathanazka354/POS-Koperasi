package transaction

import (
	"github.com/fathanazka354/pos-koperasi/internal/entity/auth"
	"github.com/fathanazka354/pos-koperasi/internal/entity/product"
	"github.com/fathanazka354/pos-koperasi/internal/entity/transaction"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/outbox"
	txdt "github.com/fathanazka354/pos-koperasi/internal/usecase/transaction/dto"
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
	Transaction transaction.Transaction `json:"transaction"`
	PaymentURL  string                  `json:"payment_url,omitempty"`
	QRString    string                  `json:"qr_string,omitempty"`
	VANumber    string                  `json:"va_number,omitempty"`

	MidtransTransactionID     string           `json:"midtrans_transaction_id,omitempty"`
	MidtransTransactionStatus string           `json:"midtrans_transaction_status,omitempty"`
	MidtransPaymentType       string           `json:"midtrans_payment_type,omitempty"`
	MidtransActions           []MidtransAction `json:"midtrans_actions,omitempty"`
	MidtransRawResponse       string           `json:"midtrans_raw_response,omitempty"`

	Message string `json:"message"`
}

type ProcessTransactionInput struct {
	Transaction transaction.Transaction
	Items       []transaction.TransactionItem
	BranchID    int
}

// MemberNotifyOutbox — payload notifikasi pembayaran; di-enqueue ke MongoDB outbox setelah settle (bukan di tx Postgres).
type MemberNotifyOutbox = outbox.MemberNotifyOutbox

type SettleInput struct {
	TransactionID int64
	BranchID      int
	MemberID      *int
	GrandTotal    float64 // metadata order (service); tidak dipakai repository untuk hitung poin
	Payment       transaction.Payment
	// RewardPoints dihitung di service/domain (RewardPointsFromPurchase); repository hanya menulis ke DB.
	RewardPoints int
}

// MidtransNotification payload bisnis untuk penyelesaian pembayaran (boundary service bebas dari paket midtrans).
type MidtransNotification = txdt.MidtransNotification

type Repository interface {
	GetProductWithStock(productID, branchID int) (*product.ProductWithStock, error)
	GetMemberByCode(memberCode string) (*auth.Member, error)
	GetMemberByID(memberID int) (*auth.Member, error)
	CreatePendingTransaction(input ProcessTransactionInput) (int64, error)
	GetTransactionByID(txID int64) (*transaction.Transaction, error)
	GetItemsByTransactionID(txID int64) ([]transaction.TransactionItem, error)
	GetTransactionsByMember(memberID int) ([]transaction.Transaction, error)
	SettleTransaction(input SettleInput) error
	CancelTransaction(transactionID int64) error
	PaymentExistsByRef(referenceNo string) (bool, error)
	GetTransactionByInvoice(invoiceNo string) (*transaction.Transaction, error)
}

type Usecase interface {
	CreateTransaction(input CreateTransactionInput, employeeID int) (*CreateTransactionOutput, error)
	HandleMidtransNotification(payload MidtransNotification) error
}

