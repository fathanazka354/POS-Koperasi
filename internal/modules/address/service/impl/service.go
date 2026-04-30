package impl

import (
	"github.com/yourname/pos-koperasi/internal/modules/address/contract"
	"github.com/yourname/pos-koperasi/internal/modules/address/domain"
)

type Service struct {
	repo contract.AddressRepository
}

func New(repo contract.AddressRepository) *Service { return &Service{repo: repo} }

var _ contract.AddressService = (*Service)(nil)

func (s *Service) Create(memberID int, input contract.CreateAddressInput) (*domain.Address, error) {
	return s.repo.Create(memberID, input)
}

func (s *Service) List(memberID int) ([]domain.Address, error) {
	return s.repo.ListByMember(memberID)
}

func (s *Service) SetDefault(id, memberID int) error {
	return s.repo.SetDefault(id, memberID)
}

func (s *Service) Delete(id, memberID int) error {
	return s.repo.Delete(id, memberID)
}
