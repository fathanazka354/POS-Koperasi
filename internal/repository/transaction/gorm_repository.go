package transaction

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	authdomain "github.com/fathanazka354/pos-koperasi/internal/entity/auth"
	productdomain "github.com/fathanazka354/pos-koperasi/internal/entity/product"
	txcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/transaction"
	txdomain "github.com/fathanazka354/pos-koperasi/internal/entity/transaction"
)

// Repository mengimplementasikan TransactionRepository menggunakan GORM.
// Query tetap berbasis SQL Postgres untuk menjaga perilaku existing.
type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ txcontract.Repository = (*Repository)(nil)

func (r *Repository) GetProductWithStock(productID, branchID int) (*productdomain.ProductWithStock, error) {
	var p productdomain.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $2
		WHERE p.id = $1 AND p.is_active = true
	`
	if err := r.db.Raw(query, productID, branchID).Scan(&p).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}
	if p.ID == 0 {
		return nil, fmt.Errorf("product not found")
	}
	return &p, nil
}

func (r *Repository) GetMemberByCode(memberCode string) (*authdomain.Member, error) {
	var m authdomain.Member
	if err := r.db.Raw("SELECT * FROM members WHERE member_code=$1 AND is_active=true", memberCode).Scan(&m).Error; err != nil {
		return nil, nil
	}
	if m.ID == 0 {
		return nil, nil
	}
	return &m, nil
}

func (r *Repository) GetMemberByID(memberID int) (*authdomain.Member, error) {
	var m authdomain.Member
	if err := r.db.Raw("SELECT * FROM members WHERE id=$1 AND is_active=true", memberID).Scan(&m).Error; err != nil {
		return nil, err
	}
	if m.ID == 0 {
		return nil, fmt.Errorf("member not found")
	}
	return &m, nil
}

func (r *Repository) CreatePendingTransaction(input txcontract.ProcessTransactionInput) (int64, error) {
	var txID int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		query := `
			INSERT INTO transactions
				(branch_id, register_id, employee_id, member_id, invoice_no,
				 subtotal, discount, tax, grand_total, status,
				 address_id, voucher_id, voucher_discount)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending',$10,$11,$12)
			RETURNING id
		`
		if err := tx.Raw(query,
			input.Transaction.BranchID,
			input.Transaction.RegisterID,
			input.Transaction.EmployeeID,
			input.Transaction.MemberID,
			input.Transaction.InvoiceNo,
			input.Transaction.Subtotal,
			input.Transaction.Discount,
			input.Transaction.Tax,
			input.Transaction.GrandTotal,
			input.Transaction.AddressID,
			input.Transaction.VoucherID,
			input.Transaction.VoucherDiscount,
		).Scan(&txID).Error; err != nil {
			return fmt.Errorf("failed to insert transaction: %w", err)
		}
		for _, item := range input.Items {
			if err := tx.Exec(`
				INSERT INTO transaction_items
					(transaction_id, product_id, quantity, unit_price, discount, subtotal)
				VALUES ($1,$2,$3,$4,$5,$6)`,
				txID, item.ProductID, item.Quantity,
				item.UnitPrice, item.Discount, item.Subtotal,
			).Error; err != nil {
				return fmt.Errorf("failed to insert item: %w", err)
			}
		}
		return nil
	})
	return txID, err
}

func (r *Repository) GetTransactionByID(txID int64) (*txdomain.Transaction, error) {
	var t txdomain.Transaction
	if err := r.db.Raw("SELECT * FROM transactions WHERE id=$1", txID).Scan(&t).Error; err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	if t.ID == 0 {
		return nil, fmt.Errorf("transaction not found")
	}
	return &t, nil
}

func (r *Repository) GetItemsByTransactionID(txID int64) ([]txdomain.TransactionItem, error) {
	var items []txdomain.TransactionItem
	if err := r.db.Raw("SELECT * FROM transaction_items WHERE transaction_id=$1", txID).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) GetTransactionsByMember(memberID int) ([]txdomain.Transaction, error) {
	var rows []txdomain.Transaction
	if err := r.db.Raw(
		`SELECT * FROM transactions WHERE member_id=$1 ORDER BY created_at DESC LIMIT 50`,
		memberID,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) SettleTransaction(input txcontract.SettleInput) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := settleMarkPaid(tx, input.TransactionID); err != nil {
			return err
		}
		if err := settleInsertPayment(tx, input); err != nil {
			return err
		}

		items, err := settleLoadItems(tx, input.TransactionID)
		if err != nil {
			return fmt.Errorf("load transaction items: %w", err)
		}
		if err := settleDeductStocks(tx, input.BranchID, items); err != nil {
			return err
		}
		if input.MemberID != nil && input.RewardPoints > 0 {
			if err := settleApplyRewardPoints(tx, *input.MemberID, input.TransactionID, input.RewardPoints); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) CancelTransaction(transactionID int64) error {
	return r.db.Exec(`
		UPDATE transactions
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, transactionID).Error
}

func (r *Repository) PaymentExistsByRef(referenceNo string) (bool, error) {
	var count int
	err := r.db.Raw("SELECT COUNT(*) FROM payments WHERE reference_no=$1", referenceNo).Scan(&count).Error
	return count > 0, err
}

func (r *Repository) GetTransactionByInvoice(invoiceNo string) (*txdomain.Transaction, error) {
	var t txdomain.Transaction
	if err := r.db.Raw("SELECT * FROM transactions WHERE invoice_no=$1", invoiceNo).Scan(&t).Error; err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	if t.ID == 0 {
		return nil, fmt.Errorf("transaction not found")
	}
	return &t, nil
}

func settleMarkPaid(tx *gorm.DB, transactionID int64) error {
	var updatedID int64
	err := tx.Raw(`
		UPDATE transactions
		SET status = 'paid', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id
	`, transactionID).Scan(&updatedID).Error
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}
	if updatedID == 0 {
		return fmt.Errorf("failed to update transaction status")
	}
	return nil
}

func settleInsertPayment(tx *gorm.DB, input txcontract.SettleInput) error {
	now := time.Now()
	if err := tx.Exec(`
		INSERT INTO payments
			(transaction_id, method, amount, change_amount, reference_no, paid_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		input.TransactionID,
		input.Payment.Method,
		input.Payment.Amount,
		input.Payment.ChangeAmount,
		input.Payment.ReferenceNo,
		&now,
	).Error; err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}
	return nil
}

func settleLoadItems(tx *gorm.DB, transactionID int64) ([]txdomain.TransactionItem, error) {
	var items []txdomain.TransactionItem
	err := tx.Raw(`SELECT * FROM transaction_items WHERE transaction_id=$1`, transactionID).Scan(&items).Error
	return items, err
}

func settleDeductStocks(tx *gorm.DB, branchID int, items []txdomain.TransactionItem) error {
	for _, item := range items {
		res := tx.Exec(`
			UPDATE stocks
			SET quantity = quantity - $1, updated_at = NOW()
			WHERE product_id = $2
			  AND branch_id  = $3
			  AND quantity   >= $1
		`, item.Quantity, item.ProductID, branchID)
		if res.Error != nil {
			return fmt.Errorf("failed to deduct stock for product %d: %w", item.ProductID, res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("insufficient stock for product_id %d", item.ProductID)
		}
	}
	return nil
}

func settleApplyRewardPoints(tx *gorm.DB, memberID int, transactionID int64, points int) error {
	if points <= 0 {
		return nil
	}
	if err := tx.Exec(`
		UPDATE members SET point_balance = point_balance + $1 WHERE id = $2
	`, points, memberID).Error; err != nil {
		return fmt.Errorf("failed to update member points: %w", err)
	}
	if err := tx.Exec(`
		INSERT INTO point_histories (member_id, transaction_id, type, points)
		VALUES ($1, $2, 'earn', $3)
	`, memberID, transactionID, points).Error; err != nil {
		return fmt.Errorf("failed to insert point history: %w", err)
	}
	return nil
}

