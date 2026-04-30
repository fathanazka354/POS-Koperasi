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

