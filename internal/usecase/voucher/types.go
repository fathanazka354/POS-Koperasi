package voucher

import "github.com/fathanazka354/pos-koperasi/internal/entity/voucher"

type ValidateResult struct {
	Voucher     *voucher.Voucher `json:"voucher"`
	DiscountAmt float64         `json:"discount_amount"`
	Label       string          `json:"label"`
}

type Repository interface {
	GetByCode(code string) (*voucher.Voucher, error)
	IncrUsed(voucherID int) error
	ListActive() ([]*voucher.Voucher, error)
}

type Usecase interface {
	Validate(code string, subtotal float64) (*ValidateResult, error)
	ListActive() ([]*voucher.Voucher, error)
}

