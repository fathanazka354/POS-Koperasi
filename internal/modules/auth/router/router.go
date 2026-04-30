package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/modules/auth/controller"
)

type Router struct {
	auth *controller.Controller
}

func New(auth *controller.Controller) *Router {
	return &Router{auth: auth}
}

func (rt *Router) Register(r chi.Router) {
	r.Post("/auth/login", rt.auth.Login)
	r.Post("/auth/member-login", rt.auth.MemberLogin)
}

