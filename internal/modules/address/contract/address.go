package contract

import "github.com/yourname/pos-koperasi/internal/modules/address/domain"

type CreateAddressInput struct {
	Label       string `json:"label"`
	Recipient   string `json:"recipient"`
	Phone       string `json:"phone"`
	AddressLine string `json:"address_line"`
	City        string `json:"city"`
	Province    string `json:"province"`
	PostalCode  string `json:"postal_code"`
	IsDefault   bool   `json:"is_default"`
}

type AddressRepository interface {
	Create(memberID int, input CreateAddressInput) (*domain.Address, error)
	ListByMember(memberID int) ([]domain.Address, error)
	GetByID(id, memberID int) (*domain.Address, error)
	SetDefault(id, memberID int) error
	Delete(id, memberID int) error
}

type AddressService interface {
	Create(memberID int, input CreateAddressInput) (*domain.Address, error)
	List(memberID int) ([]domain.Address, error)
	SetDefault(id, memberID int) error
	Delete(id, memberID int) error
}
