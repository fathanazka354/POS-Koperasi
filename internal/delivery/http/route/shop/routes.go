package shop

import (
	shophandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/shop"
	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	h *shophandler.Handler
}

func New(h *shophandler.Handler) *Routes { return &Routes{h: h} }

func (rt *Routes) Register(r chi.Router, jwtSecret string) {
	// Public routes (tanpa auth)
	r.Get("/shop/vouchers", rt.h.ListVouchers)
	r.Post("/shop/voucher/validate", rt.h.ValidateVoucher)

	// Protected member routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Post("/shop/checkout", rt.h.Checkout)
		r.Get("/shop/orders", rt.h.ListOrders)
		r.Get("/shop/orders/{invoiceNo}", rt.h.OrderDetail)
		r.Get("/shop/orders/{invoiceNo}/status", rt.h.CheckStatus)
	})
}

