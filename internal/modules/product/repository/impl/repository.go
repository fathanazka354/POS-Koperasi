package impl

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/yourname/pos-koperasi/internal/modules/product/contract"
	"github.com/yourname/pos-koperasi/internal/modules/product/domain"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

var _ contract.ProductRepository = (*Repository)(nil)

func (r *Repository) GetByBarcode(barcode string, branchID int) (*domain.ProductWithStock, error) {
	var p domain.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $2
		WHERE p.barcode = $1 AND p.is_active = true
	`
	if err := r.db.Get(&p, query, barcode, branchID); err != nil {
		return nil, fmt.Errorf("product with barcode %s not found", barcode)
	}
	return &p, nil
}

func (r *Repository) ListAll(branchID int, search string, limit, offset int) ([]domain.ProductWithStock, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []domain.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $1
		WHERE p.is_active = true AND ($2 = '' OR p.name ILIKE '%' || $2 || '%')
		ORDER BY p.name ASC
		LIMIT $3 OFFSET $4
	`
	if err := r.db.Select(&rows, query, branchID, search, limit, offset); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) GetByID(id, branchID int) (*domain.ProductWithStock, error) {
	var p domain.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $2
		WHERE p.id = $1 AND p.is_active = true
	`
	if err := r.db.Get(&p, query, id, branchID); err != nil {
		return nil, fmt.Errorf("product not found")
	}
	return &p, nil
}

func (r *Repository) ListLowStock(branchID int, limit int) ([]domain.ProductWithStock, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []domain.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $1
		WHERE p.is_active = true
		  AND COALESCE(s.quantity, 0) <= p.min_stock
		ORDER BY COALESCE(s.quantity, 0) ASC, p.name ASC
		LIMIT $2
	`
	if err := r.db.Select(&rows, query, branchID, limit); err != nil {
		return nil, err
	}
	return rows, nil
}

