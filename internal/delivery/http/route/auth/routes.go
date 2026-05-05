package auth

import (
	authhandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/auth"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	h *authhandler.Handler
}

func New(h *authhandler.Handler) *Routes {
	return &Routes{h: h}
}

func (rt *Routes) Register(r chi.Router) {
	r.Post("/auth/login", rt.h.Login)
	r.Post("/auth/member-login", rt.h.MemberLogin)
	r.Post("/auth/member-refresh", rt.h.MemberRefresh)
	r.Post("/auth/employee-refresh", rt.h.EmployeeRefresh)
}

