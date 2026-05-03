package impl

import (
	"fmt"
	"strings"

	"github.com/yourname/pos-koperasi/internal/modules/product/contract"
	"github.com/yourname/pos-koperasi/internal/modules/product/domain"
)

type Service struct {
	repo contract.ProductRepository
}

func New(repo contract.ProductRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByBarcode(barcode string, branchID int) (*domain.ProductWithStock, error) {
	return s.repo.GetByBarcode(barcode, branchID)
}

func (s *Service) ListLowStock(branchID int) ([]domain.ProductWithStock, error) {
	return s.repo.ListLowStock(branchID, 100)
}

func (s *Service) ListAll(branchID int, search string, page int) ([]domain.ProductWithStock, error) {
	const limit = 20
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	return s.repo.ListAll(branchID, search, limit, offset)
}

func (s *Service) GetByID(id, branchID int) (*domain.ProductWithStock, error) {
	return s.repo.GetByID(id, branchID)
}

func (s *Service) ListForSeller(branchID int, search string, page int) ([]domain.ProductWithStock, error) {
	const limit = 50
	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}
	return s.repo.ListForSeller(branchID, strings.TrimSpace(search), limit, offset)
}

func (s *Service) CreateForSeller(branchID int, in contract.CreateProductInput) (*domain.ProductWithStock, error) {
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
	return s.repo.CreateProduct(branchID, in)
}

func (s *Service) UpdateForSeller(branchID int, id int, in contract.UpdateProductInput) (*domain.ProductWithStock, error) {
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
	return s.repo.UpdateProduct(branchID, id, in)
}

func (s *Service) DeleteForSeller(branchID int, id int) error {
	_ = branchID // cabang tidak membatasi penghapusan logis produk global
	if id <= 0 {
		return fmt.Errorf("id tidak valid")
	}
	return s.repo.SoftDeleteProduct(id)
}

