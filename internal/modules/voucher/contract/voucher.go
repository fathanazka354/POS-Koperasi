package contract

import "github.com/yourname/pos-koperasi/internal/modules/voucher/domain"

type ValidateResult struct {
	Voucher      *domain.Voucher `json:"voucher"`
	DiscountAmt  float64         `json:"discount_amount"`
}

type VoucherRepository interface {
	GetByCode(code string) (*domain.Voucher, error)
	IncrUsed(voucherID int) error
}

type VoucherService interface {
	Validate(code string, subtotal float64) (*ValidateResult, error)
}
