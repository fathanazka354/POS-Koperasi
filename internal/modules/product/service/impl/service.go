package impl

import (
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

