package seed

import "github.com/jmoiron/sqlx"

const seedRegisterCode = "REG-001"

// CashierRegisterSeed memastikan mesin kasir awal untuk cabang.
func CashierRegisterSeed(tx *sqlx.Tx, branchID int) error {
	_, err := tx.Exec(
		`INSERT INTO cashier_registers (branch_id, register_code) VALUES ($1, $2)
		 ON CONFLICT (register_code) DO NOTHING`,
		branchID, seedRegisterCode,
	)
	return err
}
