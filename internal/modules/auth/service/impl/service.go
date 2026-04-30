package impl

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/auth/contract"
	"github.com/yourname/pos-koperasi/internal/modules/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db        *sqlx.DB
	jwtSecret string
	jwtExpiry int
}

func New(db *sqlx.DB, jwtSecret string, jwtExpiryHours string) *Service {
	expiry, _ := strconv.Atoi(jwtExpiryHours)
	if expiry == 0 {
		expiry = 24
	}
	return &Service{db: db, jwtSecret: jwtSecret, jwtExpiry: expiry}
}

func (s *Service) Login(input contract.LoginInput) (*contract.LoginOutput, error) {
	var emp domain.Employee
	err := s.db.Get(&emp,
		"SELECT * FROM employees WHERE nik=$1 AND is_active=true", input.NIK)
	if err != nil {
		return nil, fmt.Errorf("NIK atau PIN salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(emp.PinHash), []byte(input.PIN)); err != nil {
		return nil, fmt.Errorf("NIK atau PIN salah")
	}

	token, err := s.generateEmployeeToken(emp)
	if err != nil {
		return nil, fmt.Errorf("gagal generate token")
	}

	return &contract.LoginOutput{Token: token, Employee: emp}, nil
}

func (s *Service) MemberLogin(input contract.MemberLoginInput) (*contract.MemberLoginOutput, error) {
	if input.MemberCode == "" || input.Phone == "" {
		return nil, fmt.Errorf("member_code dan phone wajib diisi")
	}

	var m domain.Member
	err := s.db.Get(&m,
		`SELECT * FROM members WHERE member_code = $1 AND phone = $2 AND is_active = true`,
		input.MemberCode, input.Phone)
	if err != nil {
		return nil, fmt.Errorf("kode anggota atau nomor telepon tidak cocok")
	}

	token, err := s.generateMemberToken(m)
	if err != nil {
		return nil, fmt.Errorf("gagal generate token")
	}

	return &contract.MemberLoginOutput{Token: token, Member: m}, nil
}

func (s *Service) generateMemberToken(m domain.Member) (string, error) {
	claims := middleware.MemberClaims{
		MemberID: m.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    middleware.JWTIssuerMember,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) generateEmployeeToken(emp domain.Employee) (string, error) {
	claims := middleware.EmployeeClaims{
		EmployeeID: emp.ID,
		BranchID:   emp.BranchID,
		Role:       emp.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    middleware.JWTIssuerEmployee,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

