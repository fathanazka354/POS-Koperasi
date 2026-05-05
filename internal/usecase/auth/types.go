package auth

import "github.com/fathanazka354/pos-koperasi/internal/entity/auth"

type LoginInput struct {
	NIK string
	PIN string
}

type LoginOutput struct {
	Token    string        `json:"token"`
	Employee auth.Employee `json:"employee"`
}

type MemberLoginInput struct {
	MemberCode string
	Phone      string
}

type MemberLoginOutput struct {
	Token  string      `json:"token"`
	Member auth.Member `json:"member"`
}

type Repository interface {
	GetEmployeeByNIKActive(nik string) (*auth.Employee, error)
	GetMemberByCodeAndPhone(memberCode, phone string) (*auth.Member, error)
	GetActiveMemberByID(id int) (*auth.Member, error)
	GetActiveEmployeeByID(id int) (*auth.Employee, error)
}

type Usecase interface {
	Login(input LoginInput) (*LoginOutput, error)
	MemberLogin(input MemberLoginInput) (*MemberLoginOutput, error)
	RefreshMember(accessToken string) (*MemberLoginOutput, error)
	RefreshEmployee(accessToken string) (*LoginOutput, error)
}

