package impl

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourname/pos-koperasi/internal/modules/address/contract"
	"github.com/yourname/pos-koperasi/internal/modules/address/domain"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository { return &Repository{db: db} }

var _ contract.AddressRepository = (*Repository)(nil)

func (r *Repository) Create(memberID int, inp contract.CreateAddressInput) (*domain.Address, error) {
	if inp.IsDefault {
		_, _ = r.db.Exec(`UPDATE member_addresses SET is_default=false WHERE member_id=$1`, memberID)
	}
	var a domain.Address
	err := r.db.QueryRowx(`
		INSERT INTO member_addresses
			(member_id, label, recipient, phone, address_line, city, province, postal_code, is_default)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING *`,
		memberID, inp.Label, inp.Recipient, inp.Phone,
		inp.AddressLine, inp.City, inp.Province, inp.PostalCode, inp.IsDefault,
	).StructScan(&a)
	return &a, err
}

func (r *Repository) ListByMember(memberID int) ([]domain.Address, error) {
	var rows []domain.Address
	err := r.db.Select(&rows,
		`SELECT * FROM member_addresses WHERE member_id=$1 ORDER BY is_default DESC, id ASC`,
		memberID)
	return rows, err
}

func (r *Repository) GetByID(id, memberID int) (*domain.Address, error) {
	var a domain.Address
	err := r.db.Get(&a, `SELECT * FROM member_addresses WHERE id=$1 AND member_id=$2`, id, memberID)
	if err != nil {
		return nil, fmt.Errorf("address not found")
	}
	return &a, nil
}

func (r *Repository) SetDefault(id, memberID int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	_, err = tx.Exec(`UPDATE member_addresses SET is_default=false WHERE member_id=$1`, memberID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE member_addresses SET is_default=true WHERE id=$1 AND member_id=$2`, id, memberID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Delete(id, memberID int) error {
	_, err := r.db.Exec(`DELETE FROM member_addresses WHERE id=$1 AND member_id=$2`, id, memberID)
	return err
}
