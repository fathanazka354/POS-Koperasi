package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/controller"
)

type Router struct {
	tx *controller.Controller
}

func New(tx *controller.Controller) *Router {
	return &Router{tx: tx}
}

func (rt *Router) RegisterPublic(r chi.Router) {
	r.Post("/midtrans/webhook", rt.tx.MidtransWebhook)
}

func (rt *Router) RegisterProtected(r chi.Router) {
	r.Post("/transactions", rt.tx.CreateTransaction)
}

