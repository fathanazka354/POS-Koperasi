package product

import "github.com/fathanazka354/pos-koperasi/internal/entity/product"

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

type Repository interface {
	GetByBarcode(barcode string, branchID int) (*product.ProductWithStock, error)
	ListLowStock(branchID int, limit int) ([]product.ProductWithStock, error)
	ListAll(branchID int, search string, limit, offset int) ([]product.ProductWithStock, error)
	GetByID(id, branchID int) (*product.ProductWithStock, error)

	// Seller / manajemen produk (karyawan JWT)
	ListForSeller(branchID int, search string, limit, offset int) ([]product.ProductWithStock, error)
	CreateProduct(branchID int, in CreateProductInput) (*product.ProductWithStock, error)
	UpdateProduct(branchID int, id int, in UpdateProductInput) (*product.ProductWithStock, error)
	SoftDeleteProduct(id int) error
}

type Usecase interface {
	GetByBarcode(barcode string, branchID int) (*product.ProductWithStock, error)
	ListLowStock(branchID int) ([]product.ProductWithStock, error)
	ListAll(branchID int, search string, page int) ([]product.ProductWithStock, error)
	GetByID(id, branchID int) (*product.ProductWithStock, error)

	ListForSeller(branchID int, search string, page int) ([]product.ProductWithStock, error)
	CreateForSeller(branchID int, in CreateProductInput) (*product.ProductWithStock, error)
	UpdateForSeller(branchID int, id int, in UpdateProductInput) (*product.ProductWithStock, error)
	DeleteForSeller(branchID int, id int) error
}

