package controller

import (
	"encoding/json"
	"net/http"

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

