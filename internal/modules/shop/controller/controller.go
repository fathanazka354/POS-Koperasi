package controller

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/shop/contract"
	"github.com/yourname/pos-koperasi/pkg/response"
)

type Controller struct {
	svc contract.ShopService
}

func New(svc contract.ShopService) *Controller { return &Controller{svc: svc} }

// POST /api/v1/shop/checkout
func (c *Controller) Checkout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Login diperlukan untuk checkout")
		return
	}

	var inp contract.CheckoutInput
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
	out, err := c.svc.Checkout(inp)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Created(w, out.Message, out)
}

// GET /api/v1/shop/orders
func (c *Controller) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	txs, err := c.svc.GetTransactionsByMember(claims.MemberID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar pesanan", txs)
}

// GET /api/v1/shop/orders/{invoiceNo}
func (c *Controller) OrderDetail(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	invoiceNo := chi.URLParam(r, "invoiceNo")
	detail, err := c.svc.GetTransactionDetail(invoiceNo, claims.MemberID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	response.Success(w, "Detail pesanan", detail)
}

// GET /api/v1/shop/orders/{invoiceNo}/status
func (c *Controller) CheckStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	invoiceNo := chi.URLParam(r, "invoiceNo")
	status, err := c.svc.CheckPaymentStatus(invoiceNo, claims.MemberID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	response.Success(w, "Status pembayaran", map[string]string{"status": status, "invoice_no": invoiceNo})
}

// POST /api/v1/shop/voucher/validate  — tidak perlu login, bisa preview sebelum checkout
func (c *Controller) ValidateVoucher(w http.ResponseWriter, r *http.Request) {
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
	res, err := c.svc.ValidateVoucher(body.Code, body.Subtotal)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Voucher valid", res)
}

// GET /api/v1/shop/vouchers  — daftar voucher aktif (tanpa login)
func (c *Controller) ListVouchers(w http.ResponseWriter, r *http.Request) {
	list, err := c.svc.ListVouchers()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Daftar voucher", list)
}
