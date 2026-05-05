package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/fathanazka354/pos-koperasi/internal/entity/auth"
	"github.com/fathanazka354/pos-koperasi/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// refreshGraceAfterExpiry — selama token tidak lebih basi dari ini setelah exp, boleh diganti token baru.
const refreshGraceAfterExpiry = 7 * 24 * time.Hour

type usecase struct {
	repo      Repository
	jwtSecret string
	jwtExpiry int
}

func New(repo Repository, jwtSecret string, jwtExpiryHours string) Usecase {
	expiry, _ := strconv.Atoi(jwtExpiryHours)
	if expiry == 0 {
		expiry = 24
	}
	return &usecase{repo: repo, jwtSecret: jwtSecret, jwtExpiry: expiry}
}

func (u *usecase) Login(input LoginInput) (*LoginOutput, error) {
	emp, err := u.repo.GetEmployeeByNIKActive(input.NIK)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("NIK atau PIN salah")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(emp.PinHash), []byte(input.PIN)); err != nil {
		return nil, fmt.Errorf("NIK atau PIN salah")
	}

	token, err := u.generateEmployeeToken(*emp)
	if err != nil {
		return nil, fmt.Errorf("gagal generate token")
	}

	return &LoginOutput{Token: token, Employee: *emp}, nil
}

func (u *usecase) MemberLogin(input MemberLoginInput) (*MemberLoginOutput, error) {
	if input.MemberCode == "" || input.Phone == "" {
		return nil, fmt.Errorf("member_code dan phone wajib diisi")
	}

	m, err := u.repo.GetMemberByCodeAndPhone(input.MemberCode, input.Phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("kode anggota atau nomor telepon tidak cocok")
		}
		return nil, err
	}

	token, err := u.generateMemberToken(*m)
	if err != nil {
		return nil, fmt.Errorf("gagal generate token")
	}

	return &MemberLoginOutput{Token: token, Member: *m}, nil
}

func (u *usecase) generateMemberToken(m auth.Member) (string, error) {
	claims := model.MemberClaims{
		MemberID: m.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    model.JWTIssuerMember,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(u.jwtExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.jwtSecret))
}

func (u *usecase) generateEmployeeToken(emp auth.Employee) (string, error) {
	claims := model.EmployeeClaims{
		EmployeeID: emp.ID,
		BranchID:   emp.BranchID,
		Role:       emp.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    model.JWTIssuerEmployee,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(u.jwtExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.jwtSecret))
}

func parseMemberClaimsSkipExp(tokenStr string, secret []byte) (*model.MemberClaims, error) {
	claims := &model.MemberClaims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	_, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims.Issuer != model.JWTIssuerMember || claims.MemberID == 0 {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		if time.Since(claims.ExpiresAt.Time) > refreshGraceAfterExpiry {
			return nil, errors.New("session expired, login again")
		}
	}
	return claims, nil
}

func parseEmployeeClaimsSkipExp(tokenStr string, secret []byte) (*model.EmployeeClaims, error) {
	claims := &model.EmployeeClaims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	_, err := parser.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims.Issuer != model.JWTIssuerEmployee || claims.EmployeeID == 0 {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		if time.Since(claims.ExpiresAt.Time) > refreshGraceAfterExpiry {
			return nil, errors.New("session expired, login again")
		}
	}
	return claims, nil
}

func (u *usecase) RefreshMember(accessToken string) (*MemberLoginOutput, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("token kosong")
	}
	claims, err := parseMemberClaimsSkipExp(accessToken, []byte(u.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	m, err := u.repo.GetActiveMemberByID(claims.MemberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, err
	}

	token, err := u.generateMemberToken(*m)
	if err != nil {
		return nil, fmt.Errorf("gagal generate token")
	}
	return &MemberLoginOutput{Token: token, Member: *m}, nil
}

func (u *usecase) RefreshEmployee(accessToken string) (*LoginOutput, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("token kosong")
	}
	claims, err := parseEmployeeClaimsSkipExp(accessToken, []byte(u.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	emp, err := u.repo.GetActiveEmployeeByID(claims.EmployeeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, err
	}

	token, err := u.generateEmployeeToken(*emp)
	if err != nil {
		return nil, fmt.Errorf("gagal generate token")
	}
	return &LoginOutput{Token: token, Employee: *emp}, nil
}

