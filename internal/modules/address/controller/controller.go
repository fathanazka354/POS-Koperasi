package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/address/contract"
	"github.com/yourname/pos-koperasi/pkg/response"
)

type Controller struct {
	svc contract.AddressService
}

func New(svc contract.AddressService) *Controller { return &Controller{svc: svc} }

// GET /api/v1/member/addresses
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	items, err := c.svc.List(claims.MemberID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar alamat", items)
}

// POST /api/v1/member/addresses
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	var inp contract.CreateAddressInput
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
	addr, err := c.svc.Create(claims.MemberID, inp)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Created(w, "Alamat berhasil ditambahkan", addr)
}

// PUT /api/v1/member/addresses/{id}/default
func (c *Controller) SetDefault(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := c.svc.SetDefault(id, claims.MemberID); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Alamat default diperbarui", nil)
}

// DELETE /api/v1/member/addresses/{id}
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := c.svc.Delete(id, claims.MemberID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Alamat dihapus", nil)
}
