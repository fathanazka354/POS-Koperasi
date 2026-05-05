package auth

import (
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/fathanazka354/pos-koperasi/internal/entity/auth"
	authcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/auth"
)

// Repository mengimplementasikan AuthRepository menggunakan GORM (Postgres).
// Untuk kompatibilitas dengan service lama, error not-found dipetakan ke sql.ErrNoRows.
type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var _ authcontract.Repository = (*Repository)(nil)

func (r *Repository) GetEmployeeByNIKActive(nik string) (*auth.Employee, error) {
	var emp auth.Employee
	err := r.db.
		Table("employees").
		Where("nik = ? AND is_active = true", nik).
		Take(&emp).Error
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &emp, nil
}

func (r *Repository) GetMemberByCodeAndPhone(memberCode, phone string) (*auth.Member, error) {
	var m auth.Member
	err := r.db.
		Table("members").
		Where("member_code = ? AND phone = ? AND is_active = true", memberCode, phone).
		Take(&m).Error
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &m, nil
}

func (r *Repository) GetActiveMemberByID(id int) (*auth.Member, error) {
	var m auth.Member
	err := r.db.
		Table("members").
		Where("id = ? AND is_active = true", id).
		Take(&m).Error
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &m, nil
}

func (r *Repository) GetActiveEmployeeByID(id int) (*auth.Employee, error) {
	var emp auth.Employee
	err := r.db.
		Table("employees").
		Where("id = ? AND is_active = true", id).
		Take(&emp).Error
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &emp, nil
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sql.ErrNoRows
	}
	return err
}

