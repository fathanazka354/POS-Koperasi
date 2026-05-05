package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fathanazka354/pos-koperasi/internal/usecase/auth"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
)

type Handler struct {
	uc auth.Usecase
}

func New(uc auth.Usecase) *Handler {
	return &Handler{uc: uc}
}

type loginRequest struct {
	NIK string `json:"nik"`
	PIN string `json:"pin"`
}

type memberLoginRequest struct {
	MemberCode string `json:"member_code"`
	Phone      string `json:"phone"`
}

// POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}
	if req.NIK == "" || req.PIN == "" {
		response.BadRequest(w, "NIK dan PIN wajib diisi")
		return
	}
	out, err := h.uc.Login(auth.LoginInput{NIK: req.NIK, PIN: req.PIN})
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}
	response.Success(w, "Login berhasil", out)
}

// POST /api/v1/auth/member-login
func (h *Handler) MemberLogin(w http.ResponseWriter, r *http.Request) {
	var req memberLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}
	out, err := h.uc.MemberLogin(auth.MemberLoginInput{
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
func (h *Handler) MemberRefresh(w http.ResponseWriter, r *http.Request) {
	tok := bearerToken(r)
	if tok == "" {
		response.Unauthorized(w, "Authorization header required")
		return
	}
	out, err := h.uc.RefreshMember(tok)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}
	response.Success(w, "Token diperbarui", out)
}

// POST /api/v1/auth/employee-refresh — Authorization: Bearer <access token, boleh sudah exp>
func (h *Handler) EmployeeRefresh(w http.ResponseWriter, r *http.Request) {
	tok := bearerToken(r)
	if tok == "" {
		response.Unauthorized(w, "Authorization header required")
		return
	}
	out, err := h.uc.RefreshEmployee(tok)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}
	response.Success(w, "Token diperbarui", out)
}

