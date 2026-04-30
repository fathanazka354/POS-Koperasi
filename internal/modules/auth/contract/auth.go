package contract

import "github.com/yourname/pos-koperasi/internal/modules/auth/domain"

type LoginInput struct {
	NIK string
	PIN string
}

type LoginOutput struct {
	Token    string
	Employee domain.Employee
}

type MemberLoginInput struct {
	MemberCode string
	Phone      string
}

type MemberLoginOutput struct {
	Token  string
	Member domain.Member
}

type AuthService interface {
	Login(input LoginInput) (*LoginOutput, error)
	MemberLogin(input MemberLoginInput) (*MemberLoginOutput, error)
}

