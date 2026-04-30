package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/shop/controller"
)

type Router struct {
	ctl *controller.Controller
}

func New(ctl *controller.Controller) *Router { return &Router{ctl: ctl} }

func (rt *Router) Register(r chi.Router, jwtSecret string) {
	// Protected member routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Post("/shop/checkout", rt.ctl.Checkout)
		r.Get("/shop/orders", rt.ctl.ListOrders)
		r.Get("/shop/orders/{invoiceNo}", rt.ctl.OrderDetail)
		r.Get("/shop/orders/{invoiceNo}/status", rt.ctl.CheckStatus)
	})
}
