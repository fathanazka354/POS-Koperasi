package seed

import "gorm.io/gorm"

const seedRegisterCode = "REG-001"

// CashierRegisterSeed memastikan mesin kasir awal untuk cabang.
func CashierRegisterSeed(tx *gorm.DB, branchID int) error {
	return tx.Exec(
		`INSERT INTO cashier_registers (branch_id, register_code) VALUES ($1, $2)
		 ON CONFLICT (register_code) DO NOTHING`,
		branchID, seedRegisterCode,
	).Error
}
