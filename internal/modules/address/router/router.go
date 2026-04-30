package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/address/controller"
)

type Router struct {
	ctl *controller.Controller
}

func New(ctl *controller.Controller) *Router { return &Router{ctl: ctl} }

func (rt *Router) Register(r chi.Router, jwtSecret string) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Get("/member/addresses", rt.ctl.List)
		r.Post("/member/addresses", rt.ctl.Create)
		r.Put("/member/addresses/{id}/default", rt.ctl.SetDefault)
		r.Delete("/member/addresses/{id}", rt.ctl.Delete)
	})
}
