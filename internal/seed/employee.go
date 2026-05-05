package seed

import (
	"golang.org/x/crypto/bcrypt"

	"gorm.io/gorm"
)

const (
	seedEmployeeNIK      = "EMP001"
	seedEmployeeName     = "Kasir Satu"
	seedEmployeeRole     = "cashier"
	seedEmployeePINPlain = "1234"
)

// EmployeeSeed memastikan karyawan kasir demo (PIN: 1234).
func EmployeeSeed(tx *gorm.DB, branchID int) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(seedEmployeePINPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = tx.Exec(
		`INSERT INTO employees (branch_id, nik, full_name, role, pin_hash)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (nik) DO NOTHING`,
		branchID, seedEmployeeNIK, seedEmployeeName, seedEmployeeRole, string(hash),
	).Error
	return err
}
