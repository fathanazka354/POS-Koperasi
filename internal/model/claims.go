package model

import "github.com/golang-jwt/jwt/v5"

// Issuer JWT untuk membedakan token karyawan vs member (HTTP & WebSocket).
const (
	JWTIssuerEmployee = "pos-employee"
	JWTIssuerMember   = "pos-member"
)

// EmployeeClaims klaim JWT karyawan (POS).
type EmployeeClaims struct {
	EmployeeID int    `json:"employee_id"`
	BranchID   int    `json:"branch_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

// MemberClaims klaim JWT anggota koperasi.
type MemberClaims struct {
	MemberID int `json:"member_id"`
	jwt.RegisteredClaims
}

