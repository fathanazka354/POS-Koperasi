package contract

import "github.com/yourname/pos-koperasi/internal/modules/product/domain"

type ProductRepository interface {
	GetByBarcode(barcode string, branchID int) (*domain.ProductWithStock, error)
	ListLowStock(branchID int, limit int) ([]domain.ProductWithStock, error)
	ListAll(branchID int, search string, limit, offset int) ([]domain.ProductWithStock, error)
	GetByID(id, branchID int) (*domain.ProductWithStock, error)
}

type ProductService interface {
	GetByBarcode(barcode string, branchID int) (*domain.ProductWithStock, error)
	ListLowStock(branchID int) ([]domain.ProductWithStock, error)
	ListAll(branchID int, search string, page int) ([]domain.ProductWithStock, error)
	GetByID(id, branchID int) (*domain.ProductWithStock, error)
}

