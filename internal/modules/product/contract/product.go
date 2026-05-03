package contract

import "github.com/yourname/pos-koperasi/internal/modules/product/domain"

type ProductRepository interface {
	GetByBarcode(barcode string, branchID int) (*domain.ProductWithStock, error)
	ListLowStock(branchID int, limit int) ([]domain.ProductWithStock, error)
	ListAll(branchID int, search string, limit, offset int) ([]domain.ProductWithStock, error)
	GetByID(id, branchID int) (*domain.ProductWithStock, error)

	// Seller / manajemen produk (karyawan JWT)
	ListForSeller(branchID int, search string, limit, offset int) ([]domain.ProductWithStock, error)
	CreateProduct(branchID int, in CreateProductInput) (*domain.ProductWithStock, error)
	UpdateProduct(branchID int, id int, in UpdateProductInput) (*domain.ProductWithStock, error)
	SoftDeleteProduct(id int) error
}

// CreateProductInput input pembuatan produk cabang.
type CreateProductInput struct {
	CategoryID int
	SupplierID int
	Barcode    string
	Name       string
	Unit       string
	BuyPrice   float64
	SellPrice  float64
	MinStock   int
	Stock      int
}

// UpdateProductInput pembaruan produk + stok cabang.
type UpdateProductInput struct {
	CategoryID int
	SupplierID int
	Barcode    string
	Name       string
	Unit       string
	BuyPrice   float64
	SellPrice  float64
	MinStock   int
	Stock      int
}

type ProductService interface {
	GetByBarcode(barcode string, branchID int) (*domain.ProductWithStock, error)
	ListLowStock(branchID int) ([]domain.ProductWithStock, error)
	ListAll(branchID int, search string, page int) ([]domain.ProductWithStock, error)
	GetByID(id, branchID int) (*domain.ProductWithStock, error)

	ListForSeller(branchID int, search string, page int) ([]domain.ProductWithStock, error)
	CreateForSeller(branchID int, in CreateProductInput) (*domain.ProductWithStock, error)
	UpdateForSeller(branchID int, id int, in UpdateProductInput) (*domain.ProductWithStock, error)
	DeleteForSeller(branchID int, id int) error
}

