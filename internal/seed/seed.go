package seed

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Run menjalankan semua seeder dalam satu transaksi (idempoten / aman diulang).
func Run(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	branchID, err := BranchSeed(tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("branch: %w", err)
	}

	categories, err := CategorySeed(tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("category: %w", err)
	}

	suppliers, err := SupplierSeed(tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("supplier: %w", err)
	}

	if err := CashierRegisterSeed(tx, branchID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("cashier_register: %w", err)
	}

	if err := EmployeeSeed(tx, branchID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("employee: %w", err)
	}

	if err := ProductSeed(tx, categories, suppliers); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("product: %w", err)
	}

	if err := StockSeed(tx, branchID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("stock: %w", err)
	}

	if err := MemberSeed(tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
