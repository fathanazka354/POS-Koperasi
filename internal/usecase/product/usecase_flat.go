package product

import (
	"fmt"
	"strings"

	"github.com/fathanazka354/pos-koperasi/internal/entity/product"
)

type usecase struct {
	repo Repository
}

func New(repo Repository) Usecase { return &usecase{repo: repo} }

func (u *usecase) GetByBarcode(barcode string, branchID int) (*product.ProductWithStock, error) {
	return u.repo.GetByBarcode(barcode, branchID)
}

func (u *usecase) ListLowStock(branchID int) ([]product.ProductWithStock, error) {
	return u.repo.ListLowStock(branchID, 100)
}

func (u *usecase) ListAll(branchID int, search string, page int) ([]product.ProductWithStock, error) {
	const limit = 20
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	return u.repo.ListAll(branchID, search, limit, offset)
}

func (u *usecase) GetByID(id, branchID int) (*product.ProductWithStock, error) {
	return u.repo.GetByID(id, branchID)
}

func (u *usecase) ListForSeller(branchID int, search string, page int) ([]product.ProductWithStock, error) {
	const limit = 50
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	return u.repo.ListForSeller(branchID, strings.TrimSpace(search), limit, offset)
}

func (u *usecase) CreateForSeller(branchID int, in CreateProductInput) (*product.ProductWithStock, error) {
	in.Barcode = strings.TrimSpace(in.Barcode)
	in.Name = strings.TrimSpace(in.Name)
	in.Unit = strings.TrimSpace(in.Unit)
	if in.Unit == "" {
		in.Unit = "pcs"
	}
	if in.CategoryID <= 0 {
		in.CategoryID = 1
	}
	if in.SupplierID <= 0 {
		in.SupplierID = 1
	}
	if in.Barcode == "" || in.Name == "" {
		return nil, fmt.Errorf("barcode dan nama wajib diisi")
	}
	if in.BuyPrice <= 0 || in.SellPrice <= 0 {
		return nil, fmt.Errorf("harga beli dan jual harus lebih dari 0")
	}
	if in.MinStock < 0 {
		in.MinStock = 0
	}
	if in.Stock < 0 {
		in.Stock = 0
	}
	return u.repo.CreateProduct(branchID, in)
}

func (u *usecase) UpdateForSeller(branchID int, id int, in UpdateProductInput) (*product.ProductWithStock, error) {
	in.Barcode = strings.TrimSpace(in.Barcode)
	in.Name = strings.TrimSpace(in.Name)
	in.Unit = strings.TrimSpace(in.Unit)
	if in.Unit == "" {
		in.Unit = "pcs"
	}
	if in.CategoryID <= 0 {
		in.CategoryID = 1
	}
	if in.SupplierID <= 0 {
		in.SupplierID = 1
	}
	if id <= 0 || in.Barcode == "" || in.Name == "" {
		return nil, fmt.Errorf("data produk tidak valid")
	}
	if in.BuyPrice <= 0 || in.SellPrice <= 0 {
		return nil, fmt.Errorf("harga beli dan jual harus lebih dari 0")
	}
	if in.MinStock < 0 {
		in.MinStock = 0
	}
	if in.Stock < 0 {
		in.Stock = 0
	}
	return u.repo.UpdateProduct(branchID, id, in)
}

func (u *usecase) DeleteForSeller(branchID int, id int) error {
	_ = branchID // cabang tidak membatasi penghapusan logis produk global
	if id <= 0 {
		return fmt.Errorf("id tidak valid")
	}
	return u.repo.SoftDeleteProduct(id)
}

