package transaction

import (
	txhandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/transaction"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	h *txhandler.Handler
}

func New(h *txhandler.Handler) *Routes { return &Routes{h: h} }

func (rt *Routes) RegisterPublic(r chi.Router) {
	r.Post("/midtrans/webhook", rt.h.MidtransWebhook)
}

func (rt *Routes) RegisterProtected(r chi.Router) {
	r.Post("/transactions", rt.h.CreateTransaction)
}

