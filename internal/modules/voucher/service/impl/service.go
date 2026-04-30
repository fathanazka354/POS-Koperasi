package impl

import (
	"fmt"
	"math"
	"time"

	"github.com/yourname/pos-koperasi/internal/modules/voucher/contract"
)

type Service struct {
	repo contract.VoucherRepository
}

func New(repo contract.VoucherRepository) *Service { return &Service{repo: repo} }

var _ contract.VoucherService = (*Service)(nil)

func (s *Service) Validate(code string, subtotal float64) (*contract.ValidateResult, error) {
	v, err := s.repo.GetByCode(code)
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
	switch v.DiscountType {
	case "percent":
		discount = math.Round(subtotal * v.Value / 100)
		if v.MaxDiscount != nil && discount > *v.MaxDiscount {
			discount = *v.MaxDiscount
		}
	case "fixed":
		discount = v.Value
	}

	return &contract.ValidateResult{Voucher: v, DiscountAmt: discount}, nil
}
