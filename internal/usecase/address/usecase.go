package address

import "github.com/fathanazka354/pos-koperasi/internal/entity/address"

type usecase struct {
	repo Repository
}

func New(repo Repository) Usecase { return &usecase{repo: repo} }

func (u *usecase) Create(memberID int, input CreateAddressInput) (*address.Address, error) {
	return u.repo.Create(memberID, input)
}

func (u *usecase) List(memberID int) ([]address.Address, error) {
	return u.repo.ListByMember(memberID)
}

func (u *usecase) SetDefault(id, memberID int) error {
	return u.repo.SetDefault(id, memberID)
}

func (u *usecase) Delete(id, memberID int) error {
	return u.repo.Delete(id, memberID)
}
