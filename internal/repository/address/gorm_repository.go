package address

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/fathanazka354/pos-koperasi/internal/entity/address"
	addrcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/address"
)

// Repository mengimplementasikan AddressRepository menggunakan GORM.
type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ addrcontract.Repository = (*Repository)(nil)

func (r *Repository) Create(memberID int, inp addrcontract.CreateAddressInput) (*address.Address, error) {
	var out address.Address
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if inp.IsDefault {
			if err := tx.Exec(`UPDATE member_addresses SET is_default=false WHERE member_id=$1`, memberID).Error; err != nil {
				return err
			}
		}
		if err := tx.Raw(`
			INSERT INTO member_addresses
				(member_id, label, recipient, phone, address_line, city, province, postal_code, is_default)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING *`,
			memberID, inp.Label, inp.Recipient, inp.Phone,
			inp.AddressLine, inp.City, inp.Province, inp.PostalCode, inp.IsDefault,
		).Scan(&out).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *Repository) ListByMember(memberID int) ([]address.Address, error) {
	var rows []address.Address
	err := r.db.Raw(
		`SELECT * FROM member_addresses WHERE member_id=$1 ORDER BY is_default DESC, id ASC`,
		memberID,
	).Scan(&rows).Error
	return rows, err
}

func (r *Repository) GetByID(id, memberID int) (*address.Address, error) {
	var a address.Address
	err := r.db.Raw(`SELECT * FROM member_addresses WHERE id=$1 AND member_id=$2`, id, memberID).Scan(&a).Error
	if err != nil {
		return nil, err
	}
	if a.ID == 0 {
		return nil, fmt.Errorf("address not found")
	}
	return &a, nil
}

func (r *Repository) SetDefault(id, memberID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`UPDATE member_addresses SET is_default=false WHERE member_id=$1`, memberID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE member_addresses SET is_default=true WHERE id=$1 AND member_id=$2`, id, memberID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) Delete(id, memberID int) error {
	return r.db.Exec(`DELETE FROM member_addresses WHERE id=$1 AND member_id=$2`, id, memberID).Error
}

