package impl

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	authdomain "github.com/yourname/pos-koperasi/internal/modules/auth/domain"
	productdomain "github.com/yourname/pos-koperasi/internal/modules/product/domain"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/domain"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

var _ contract.TransactionRepository = (*Repository)(nil)

func (r *Repository) GetProductWithStock(productID, branchID int) (*productdomain.ProductWithStock, error) {
	var p productdomain.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $2
		WHERE p.id = $1 AND p.is_active = true
	`
	if err := r.db.Get(&p, query, productID, branchID); err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}
	return &p, nil
}

func (r *Repository) GetMemberByCode(memberCode string) (*authdomain.Member, error) {
	var m authdomain.Member
	if err := r.db.Get(&m, "SELECT * FROM members WHERE member_code=$1 AND is_active=true", memberCode); err != nil {
		return nil, nil
	}
	return &m, nil
}

func (r *Repository) GetTransactionByInvoice(invoiceNo string) (*domain.Transaction, error) {
	var t domain.Transaction
	if err := r.db.Get(&t, "SELECT * FROM transactions WHERE invoice_no=$1", invoiceNo); err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	return &t, nil
}

func (r *Repository) GetTransactionByID(txID int64) (*domain.Transaction, error) {
	var t domain.Transaction
	if err := r.db.Get(&t, "SELECT * FROM transactions WHERE id=$1", txID); err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	return &t, nil
}

func (r *Repository) PaymentExistsByRef(referenceNo string) (bool, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM payments WHERE reference_no=$1", referenceNo)
	return count > 0, err
}

func (r *Repository) GetItemsByTransactionID(txID int64) ([]domain.TransactionItem, error) {
	var items []domain.TransactionItem
	if err := r.db.Select(&items, "SELECT * FROM transaction_items WHERE transaction_id=$1", txID); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) CreatePendingTransaction(input contract.ProcessTransactionInput) (int64, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var txID int64
	query := `
		INSERT INTO transactions
			(branch_id, register_id, employee_id, member_id, invoice_no,
			 subtotal, discount, tax, grand_total, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending')
		RETURNING id
	`
	err = tx.QueryRow(query,
		input.Transaction.BranchID,
		input.Transaction.RegisterID,
		input.Transaction.EmployeeID,
		input.Transaction.MemberID,
		input.Transaction.InvoiceNo,
		input.Transaction.Subtotal,
		input.Transaction.Discount,
		input.Transaction.Tax,
		input.Transaction.GrandTotal,
	).Scan(&txID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert transaction: %w", err)
	}

	for _, item := range input.Items {
		_, err = tx.Exec(`
			INSERT INTO transaction_items
				(transaction_id, product_id, quantity, unit_price, discount, subtotal)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			txID, item.ProductID, item.Quantity,
			item.UnitPrice, item.Discount, item.Subtotal,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to insert item: %w", err)
		}
	}

	return txID, tx.Commit()
}

func (r *Repository) SettleTransaction(input contract.SettleInput) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var updatedID int64
	err = tx.QueryRow(`
		UPDATE transactions
		SET status = 'paid', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id
	`, input.TransactionID).Scan(&updatedID)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO payments
			(transaction_id, method, amount, change_amount, reference_no, paid_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		input.TransactionID,
		input.Payment.Method,
		input.Payment.Amount,
		input.Payment.ChangeAmount,
		input.Payment.ReferenceNo,
		&now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	items, err := r.GetItemsByTransactionID(input.TransactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction items: %w", err)
	}

	for _, item := range items {
		result, err := tx.Exec(`
			UPDATE stocks
			SET quantity = quantity - $1, updated_at = NOW()
			WHERE product_id = $2
			  AND branch_id  = $3
			  AND quantity   >= $1
		`, item.Quantity, item.ProductID, input.BranchID)
		if err != nil {
			return fmt.Errorf("failed to deduct stock for product %d: %w", item.ProductID, err)
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("insufficient stock for product_id %d", item.ProductID)
		}
	}

	if input.MemberID != nil {
		earnedPoints := int(input.GrandTotal / 10000)
		if earnedPoints > 0 {
			_, err = tx.Exec(`
				UPDATE members SET point_balance = point_balance + $1 WHERE id = $2
			`, earnedPoints, *input.MemberID)
			if err != nil {
				return fmt.Errorf("failed to update member points: %w", err)
			}

			_, err = tx.Exec(`
				INSERT INTO point_histories (member_id, transaction_id, type, points)
				VALUES ($1, $2, 'earn', $3)
			`, *input.MemberID, input.TransactionID, earnedPoints)
			if err != nil {
				return fmt.Errorf("failed to insert point history: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (r *Repository) CancelTransaction(transactionID int64) error {
	_, err := r.db.Exec(`
		UPDATE transactions
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, transactionID)
	return err
}

