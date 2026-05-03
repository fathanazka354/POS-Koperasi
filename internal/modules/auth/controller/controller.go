package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourname/pos-koperasi/internal/modules/auth/contract"
	"github.com/yourname/pos-koperasi/internal/modules/auth/controller/dto"
	"github.com/yourname/pos-koperasi/pkg/response"
)

type Controller struct {
	svc contract.AuthService
}

func New(svc contract.AuthService) *Controller {
	return &Controller{svc: svc}
}

// POST /api/v1/auth/login
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.NIK == "" || req.PIN == "" {
		response.BadRequest(w, "NIK dan PIN wajib diisi")
		return
	}

	out, err := c.svc.Login(contract.LoginInput{NIK: req.NIK, PIN: req.PIN})
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.Success(w, "Login berhasil", out)
}

// POST /api/v1/auth/member-login
func (c *Controller) MemberLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.MemberLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	out, err := c.svc.MemberLogin(contract.MemberLoginInput{
		MemberCode: req.MemberCode,
		Phone:      req.Phone,
	})
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.Success(w, "Login member berhasil", out)
}

func bearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

// POST /api/v1/auth/member-refresh — Authorization: Bearer <access token, boleh sudah exp>
func (c *Controller) MemberRefresh(w http.ResponseWriter, r *http.Request) {
	tok := bearerToken(r)
	if tok == "" {
		response.Unauthorized(w, "Authorization header required")
		return
	}
	out, err := c.svc.RefreshMember(tok)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}
	response.Success(w, "Token diperbarui", out)
}

// POST /api/v1/auth/employee-refresh — Authorization: Bearer <access token, boleh sudah exp>
func (c *Controller) EmployeeRefresh(w http.ResponseWriter, r *http.Request) {
	tok := bearerToken(r)
	if tok == "" {
		response.Unauthorized(w, "Authorization header required")
		return
	}
	out, err := c.svc.RefreshEmployee(tok)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}
	response.Success(w, "Token diperbarui", out)
}

