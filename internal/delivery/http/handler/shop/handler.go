package shop

import (
	"encoding/json"
	"net/http"

	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/shop"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	uc shop.Usecase
}

func New(uc shop.Usecase) *Handler {
	return &Handler{uc: uc}
}

// POST /api/v1/shop/checkout
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Login diperlukan untuk checkout")
		return
	}

	var inp shop.CheckoutInput
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}
	if len(inp.Items) == 0 {
		response.BadRequest(w, "Keranjang belanja kosong")
		return
	}
	if inp.PayMethod == "" {
		response.BadRequest(w, "Metode pembayaran wajib diisi")
		return
	}

	inp.MemberID = claims.MemberID
	out, err := h.uc.Checkout(inp)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, out.Message, out)
}

// GET /api/v1/shop/orders
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	txs, err := h.uc.GetTransactionsByMember(claims.MemberID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar pesanan", txs)
}

// GET /api/v1/shop/orders/{invoiceNo}
func (h *Handler) OrderDetail(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	invoiceNo := chi.URLParam(r, "invoiceNo")
	detail, err := h.uc.GetTransactionDetail(invoiceNo, claims.MemberID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	response.Success(w, "Detail pesanan", detail)
}

// GET /api/v1/shop/orders/{invoiceNo}/status
func (h *Handler) CheckStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	invoiceNo := chi.URLParam(r, "invoiceNo")
	status, err := h.uc.CheckPaymentStatus(invoiceNo, claims.MemberID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	response.Success(w, "Status pembayaran", map[string]string{"status": status, "invoice_no": invoiceNo})
}

// POST /api/v1/shop/voucher/validate
func (h *Handler) ValidateVoucher(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string  `json:"code"`
		Subtotal float64 `json:"subtotal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}
	if body.Code == "" {
		response.BadRequest(w, "Kode voucher tidak boleh kosong")
		return
	}
	res, err := h.uc.ValidateVoucher(body.Code, body.Subtotal)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Voucher valid", res)
}

// GET /api/v1/shop/vouchers
func (h *Handler) ListVouchers(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.ListVouchers()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar voucher", list)
}

