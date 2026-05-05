package address

import "github.com/fathanazka354/pos-koperasi/internal/entity/address"

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

type Repository interface {
	Create(memberID int, input CreateAddressInput) (*address.Address, error)
	ListByMember(memberID int) ([]address.Address, error)
	GetByID(id, memberID int) (*address.Address, error)
	SetDefault(id, memberID int) error
	Delete(id, memberID int) error
}

type Usecase interface {
	Create(memberID int, input CreateAddressInput) (*address.Address, error)
	List(memberID int) ([]address.Address, error)
	SetDefault(id, memberID int) error
	Delete(id, memberID int) error
}

