package address

import (
	"github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/address"
	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	h *address.Handler
}

func New(h *address.Handler) *Routes { return &Routes{h: h} }

func (rt *Routes) Register(r chi.Router, jwtSecret string) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Get("/member/addresses", rt.h.List)
		r.Post("/member/addresses", rt.h.Create)
		r.Put("/member/addresses/{id}/default", rt.h.SetDefault)
		r.Delete("/member/addresses/{id}", rt.h.Delete)
	})
}

