package impl

import (
	"database/sql"
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

// ListForSeller — daftar produk aktif untuk manajemen penjual.
func (r *Repository) ListForSeller(branchID int, search string, limit, offset int) ([]domain.ProductWithStock, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var rows []domain.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $1
		WHERE p.is_active = true
		  AND ($2 = '' OR p.name ILIKE '%' || $2 || '%' OR p.barcode ILIKE '%' || $2 || '%')
		ORDER BY p.id DESC
		LIMIT $3 OFFSET $4
	`
	if err := r.db.Select(&rows, query, branchID, search, limit, offset); err != nil {
		return nil, err
	}
	return rows, nil
}

// CreateProduct membuat produk + stok cabang; reaktivasi jika barcode tidak aktif.
func (r *Repository) CreateProduct(branchID int, in contract.CreateProductInput) (*domain.ProductWithStock, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var existingID int
	var wasActive bool
	err = tx.QueryRow(`SELECT id, is_active FROM products WHERE barcode = $1`, in.Barcode).Scan(&existingID, &wasActive)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil && existingID > 0 {
		if wasActive {
			return nil, fmt.Errorf("barcode %s sudah dipakai", in.Barcode)
		}
		_, err = tx.Exec(`
			UPDATE products SET category_id=$1, supplier_id=$2, name=$3, unit=$4,
				buy_price=$5, sell_price=$6, min_stock=$7, is_active=true WHERE id=$8`,
			in.CategoryID, in.SupplierID, in.Name, in.Unit, in.BuyPrice, in.SellPrice, in.MinStock, existingID)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(`
			INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
			ON CONFLICT (product_id, branch_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
			existingID, branchID, in.Stock)
		if err != nil {
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return r.GetByID(existingID, branchID)
	}

	var id int
	err = tx.QueryRow(`
		INSERT INTO products (category_id, supplier_id, barcode, name, unit, buy_price, sell_price, min_stock, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
		RETURNING id`,
		in.CategoryID, in.SupplierID, in.Barcode, in.Name, in.Unit, in.BuyPrice, in.SellPrice, in.MinStock).Scan(&id)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`
		INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
		ON CONFLICT (product_id, branch_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
		id, branchID, in.Stock)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(id, branchID)
}

// UpdateProduct memperbarui produk dan kuantitas stok cabang.
func (r *Repository) UpdateProduct(branchID int, id int, in contract.UpdateProductInput) (*domain.ProductWithStock, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		UPDATE products SET category_id=$1, supplier_id=$2, barcode=$3, name=$4, unit=$5,
			buy_price=$6, sell_price=$7, min_stock=$8 WHERE id=$9 AND is_active=true`,
		in.CategoryID, in.SupplierID, in.Barcode, in.Name, in.Unit,
		in.BuyPrice, in.SellPrice, in.MinStock, id)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, fmt.Errorf("produk tidak ditemukan atau tidak aktif")
	}
	_, err = tx.Exec(`
		INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
		ON CONFLICT (product_id, branch_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
		id, branchID, in.Stock)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(id, branchID)
}

// SoftDeleteProduct menonaktifkan produk (tidak dihapus fisik).
func (r *Repository) SoftDeleteProduct(id int) error {
	res, err := r.db.Exec(`UPDATE products SET is_active = false WHERE id = $1 AND is_active = true`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("produk tidak ditemukan atau sudah tidak aktif")
	}
	return nil
}

