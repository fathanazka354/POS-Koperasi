package impl

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourname/pos-koperasi/internal/modules/voucher/contract"
	"github.com/yourname/pos-koperasi/internal/modules/voucher/domain"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository { return &Repository{db: db} }

var _ contract.VoucherRepository = (*Repository)(nil)

func (r *Repository) GetByCode(code string) (*domain.Voucher, error) {
	var v domain.Voucher
	err := r.db.Get(&v, `SELECT * FROM vouchers WHERE code=$1`, code)
	if err != nil {
		return nil, fmt.Errorf("voucher tidak ditemukan")
	}
	return &v, nil
}

func (r *Repository) IncrUsed(voucherID int) error {
	_, err := r.db.Exec(`UPDATE vouchers SET used_count=used_count+1 WHERE id=$1`, voucherID)
	return err
}

func (r *Repository) ListActive() ([]*domain.Voucher, error) {
	var list []*domain.Voucher
	err := r.db.Select(&list, `
		SELECT * FROM vouchers
		WHERE is_active = true
		  AND start_date <= CURRENT_DATE
		  AND end_date   >= CURRENT_DATE
		  AND (quota IS NULL OR used_count < quota)
		ORDER BY discount_type, id
	`)
	if err != nil {
		return nil, err
	}
	return list, nil
}
