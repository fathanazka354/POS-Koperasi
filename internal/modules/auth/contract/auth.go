package contract

import "github.com/yourname/pos-koperasi/internal/modules/auth/domain"

type LoginInput struct {
	NIK string
	PIN string
}

type LoginOutput struct {
	Token    string          `json:"token"`
	Employee domain.Employee `json:"employee"`
}

type MemberLoginInput struct {
	MemberCode string
	Phone      string
}

type MemberLoginOutput struct {
	Token  string        `json:"token"`
	Member domain.Member `json:"member"`
}

type AuthService interface {
	Login(input LoginInput) (*LoginOutput, error)
	MemberLogin(input MemberLoginInput) (*MemberLoginOutput, error)
}

