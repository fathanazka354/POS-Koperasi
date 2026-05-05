package voucher

import (
	"fmt"
	"math"
	"time"

	"github.com/fathanazka354/pos-koperasi/internal/entity/voucher"
)

type usecase struct {
	repo Repository
}

func New(repo Repository) Usecase { return &usecase{repo: repo} }

func (u *usecase) Validate(code string, subtotal float64) (*ValidateResult, error) {
	v, err := u.repo.GetByCode(code)
	if err != nil {
		return nil, err
	}
	if !v.IsActive {
		return nil, fmt.Errorf("voucher tidak aktif")
	}
	now := time.Now()
	if now.Before(v.StartDate) || now.After(v.EndDate) {
		return nil, fmt.Errorf("voucher sudah kadaluarsa atau belum berlaku")
	}
	if v.UsedCount >= v.Quota {
		return nil, fmt.Errorf("kuota voucher sudah habis")
	}
	if subtotal < v.MinPurchase {
		return nil, fmt.Errorf("minimum pembelian Rp %.0f untuk menggunakan voucher ini", v.MinPurchase)
	}

	var discount float64
	var label string
	switch v.DiscountType {
	case "percent":
		discount = math.Round(subtotal * v.Value / 100)
		if v.MaxDiscount != nil && discount > *v.MaxDiscount {
			discount = *v.MaxDiscount
		}
		label = fmt.Sprintf("Diskon %.0f%%", v.Value)
		if v.MaxDiscount != nil {
			label += fmt.Sprintf(" (maks Rp %.0f)", *v.MaxDiscount)
		}
	case "fixed":
		discount = v.Value
		label = fmt.Sprintf("Diskon Rp %.0f", v.Value)
	case "ongkir":
		discount = v.Value
		label = fmt.Sprintf("Gratis Ongkir Rp %.0f", v.Value)
	}

	return &ValidateResult{Voucher: v, DiscountAmt: discount, Label: label}, nil
}

func (u *usecase) ListActive() ([]*voucher.Voucher, error) {
	return u.repo.ListActive()
}

