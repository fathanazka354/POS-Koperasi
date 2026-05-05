package address

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/address"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	uc address.Usecase
}

func New(uc address.Usecase) *Handler {
	return &Handler{uc: uc}
}

// GET /api/v1/member/addresses
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	items, err := h.uc.List(claims.MemberID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar alamat", items)
}

// POST /api/v1/member/addresses
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	var inp address.CreateAddressInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}
	if inp.Recipient == "" || inp.AddressLine == "" || inp.City == "" {
		response.BadRequest(w, "recipient, address_line, dan city wajib diisi")
		return
	}
	if inp.Label == "" {
		inp.Label = "Rumah"
	}
	addr, err := h.uc.Create(claims.MemberID, inp)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Created(w, "Alamat berhasil ditambahkan", addr)
}

// PUT /api/v1/member/addresses/{id}/default
func (h *Handler) SetDefault(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.uc.SetDefault(id, claims.MemberID); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Alamat default diperbarui", nil)
}

// DELETE /api/v1/member/addresses/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.uc.Delete(id, claims.MemberID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Alamat dihapus", nil)
}

