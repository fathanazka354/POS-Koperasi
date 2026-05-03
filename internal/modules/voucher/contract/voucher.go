package contract

import "github.com/yourname/pos-koperasi/internal/modules/voucher/domain"

type ValidateResult struct {
	Voucher     *domain.Voucher `json:"voucher"`
	DiscountAmt float64         `json:"discount_amount"`
	Label       string          `json:"label"`
}

type VoucherRepository interface {
	GetByCode(code string) (*domain.Voucher, error)
	IncrUsed(voucherID int) error
	ListActive() ([]*domain.Voucher, error)
}

type VoucherService interface {
	Validate(code string, subtotal float64) (*ValidateResult, error)
	ListActive() ([]*domain.Voucher, error)
}
