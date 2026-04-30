package dto

type LoginRequest struct {
	NIK string `json:"nik"`
	PIN string `json:"pin"`
}

type MemberLoginRequest struct {
	MemberCode string `json:"member_code"`
	Phone      string `json:"phone"`
}

