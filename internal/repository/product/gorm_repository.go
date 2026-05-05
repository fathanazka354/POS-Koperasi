package product

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	"github.com/fathanazka354/pos-koperasi/internal/entity/product"
	productcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/product"
)

// Repository mengimplementasikan ProductRepository menggunakan GORM.
// Query tetap menggunakan SQL Postgres untuk menjaga perilaku/hasil join (ProductWithStock).
type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var _ productcontract.Repository = (*Repository)(nil)

func (r *Repository) GetByBarcode(barcode string, branchID int) (*product.ProductWithStock, error) {
	var p product.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $2
		WHERE p.barcode = $1 AND p.is_active = true
	`
	if err := r.db.Raw(query, barcode, branchID).Scan(&p).Error; err != nil {
		return nil, err
	}
	// Scan tanpa error bisa tetap menghasilkan zero struct; guard minimal:
	if p.ID == 0 {
		return nil, fmt.Errorf("product with barcode %s not found", barcode)
	}
	return &p, nil
}

func (r *Repository) ListLowStock(branchID int, limit int) ([]product.ProductWithStock, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []product.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $1
		WHERE p.is_active = true
		  AND COALESCE(s.quantity, 0) <= p.min_stock
		ORDER BY COALESCE(s.quantity, 0) ASC, p.name ASC
		LIMIT $2
	`
	if err := r.db.Raw(query, branchID, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) ListAll(branchID int, search string, limit, offset int) ([]product.ProductWithStock, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var rows []product.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $1
		WHERE p.is_active = true AND ($2 = '' OR p.name ILIKE '%' || $2 || '%')
		ORDER BY p.name ASC
		LIMIT $3 OFFSET $4
	`
	if err := r.db.Raw(query, branchID, search, limit, offset).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) GetByID(id, branchID int) (*product.ProductWithStock, error) {
	var p product.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $2
		WHERE p.id = $1 AND p.is_active = true
	`
	if err := r.db.Raw(query, id, branchID).Scan(&p).Error; err != nil {
		return nil, err
	}
	if p.ID == 0 {
		return nil, fmt.Errorf("product not found")
	}
	return &p, nil
}

func (r *Repository) ListForSeller(branchID int, search string, limit, offset int) ([]product.ProductWithStock, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var rows []product.ProductWithStock
	query := `
		SELECT p.*, COALESCE(s.quantity, 0) AS stock
		FROM products p
		LEFT JOIN stocks s ON s.product_id = p.id AND s.branch_id = $1
		WHERE p.is_active = true
		  AND ($2 = '' OR p.name ILIKE '%' || $2 || '%' OR p.barcode ILIKE '%' || $2 || '%')
		ORDER BY p.id DESC
		LIMIT $3 OFFSET $4
	`
	if err := r.db.Raw(query, branchID, search, limit, offset).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) CreateProduct(branchID int, in productcontract.CreateProductInput) (*product.ProductWithStock, error) {
	var out *product.ProductWithStock
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existingID int
		var wasActive bool
		row := tx.Raw(`SELECT id, is_active FROM products WHERE barcode = $1`, in.Barcode).Row()
		if err := row.Scan(&existingID, &wasActive); err != nil && err != sql.ErrNoRows {
			return err
		}

		if existingID > 0 {
			if wasActive {
				return fmt.Errorf("barcode %s sudah dipakai", in.Barcode)
			}
			if err := tx.Exec(`
				UPDATE products SET category_id=$1, supplier_id=$2, name=$3, unit=$4,
					buy_price=$5, sell_price=$6, min_stock=$7, is_active=true WHERE id=$8`,
				in.CategoryID, in.SupplierID, in.Name, in.Unit, in.BuyPrice, in.SellPrice, in.MinStock, existingID,
			).Error; err != nil {
				return err
			}
			if err := tx.Exec(`
				INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
				ON CONFLICT (product_id, branch_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
				existingID, branchID, in.Stock,
			).Error; err != nil {
				return err
			}
			p, err := r.GetByID(existingID, branchID)
			if err != nil {
				return err
			}
			out = p
			return nil
		}

		var id int
		if err := tx.Raw(`
			INSERT INTO products (category_id, supplier_id, barcode, name, unit, buy_price, sell_price, min_stock, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true)
			RETURNING id`,
			in.CategoryID, in.SupplierID, in.Barcode, in.Name, in.Unit, in.BuyPrice, in.SellPrice, in.MinStock,
		).Scan(&id).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
			ON CONFLICT (product_id, branch_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
			id, branchID, in.Stock,
		).Error; err != nil {
			return err
		}
		p, err := r.GetByID(id, branchID)
		if err != nil {
			return err
		}
		out = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) UpdateProduct(branchID int, id int, in productcontract.UpdateProductInput) (*product.ProductWithStock, error) {
	var out *product.ProductWithStock
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Exec(`
			UPDATE products SET category_id=$1, supplier_id=$2, barcode=$3, name=$4, unit=$5,
				buy_price=$6, sell_price=$7, min_stock=$8 WHERE id=$9 AND is_active=true`,
			in.CategoryID, in.SupplierID, in.Barcode, in.Name, in.Unit,
			in.BuyPrice, in.SellPrice, in.MinStock, id,
		)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("produk tidak ditemukan atau tidak aktif")
		}
		if err := tx.Exec(`
			INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
			ON CONFLICT (product_id, branch_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
			id, branchID, in.Stock,
		).Error; err != nil {
			return err
		}
		p, err := r.GetByID(id, branchID)
		if err != nil {
			return err
		}
		out = p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) SoftDeleteProduct(id int) error {
	res := r.db.Exec(`UPDATE products SET is_active = false WHERE id = $1 AND is_active = true`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("produk tidak ditemukan atau sudah tidak aktif")
	}
	return nil
}

