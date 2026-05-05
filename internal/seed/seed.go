package seed

import (
	"fmt"

	"gorm.io/gorm"
)

// Run menjalankan semua seeder dalam satu transaksi (idempoten / aman diulang).
func Run(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		branchID, err := BranchSeed(tx)
		if err != nil {
			return fmt.Errorf("branch: %w", err)
		}

		categories, err := CategorySeed(tx)
		if err != nil {
			return fmt.Errorf("category: %w", err)
		}

		suppliers, err := SupplierSeed(tx)
		if err != nil {
			return fmt.Errorf("supplier: %w", err)
		}

		if err := CashierRegisterSeed(tx, branchID); err != nil {
			return fmt.Errorf("cashier_register: %w", err)
		}

		if err := EmployeeSeed(tx, branchID); err != nil {
			return fmt.Errorf("employee: %w", err)
		}

		if err := ProductSeed(tx, categories, suppliers); err != nil {
			return fmt.Errorf("product: %w", err)
		}

		if err := StockSeed(tx, branchID); err != nil {
			return fmt.Errorf("stock: %w", err)
		}

		if err := MemberSeed(tx); err != nil {
			return fmt.Errorf("member: %w", err)
		}

		return nil
	})
}
