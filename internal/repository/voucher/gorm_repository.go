package voucher

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/fathanazka354/pos-koperasi/internal/entity/voucher"
	vouchercontract "github.com/fathanazka354/pos-koperasi/internal/usecase/voucher"
)

// Repository mengimplementasikan VoucherRepository menggunakan GORM.
type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ vouchercontract.Repository = (*Repository)(nil)

func (r *Repository) GetByCode(code string) (*voucher.Voucher, error) {
	var v voucher.Voucher
	if err := r.db.Raw(`SELECT * FROM vouchers WHERE code=$1`, code).Scan(&v).Error; err != nil {
		return nil, err
	}
	if v.ID == 0 {
		return nil, fmt.Errorf("voucher tidak ditemukan")
	}
	return &v, nil
}

func (r *Repository) IncrUsed(voucherID int) error {
	return r.db.Exec(`UPDATE vouchers SET used_count=used_count+1 WHERE id=$1`, voucherID).Error
}

func (r *Repository) ListActive() ([]*voucher.Voucher, error) {
	var list []*voucher.Voucher
	err := r.db.Raw(`
		SELECT * FROM vouchers
		WHERE is_active = true
		  AND start_date <= CURRENT_DATE
		  AND end_date   >= CURRENT_DATE
		  AND (quota IS NULL OR used_count < quota)
		ORDER BY discount_type, id
	`).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

