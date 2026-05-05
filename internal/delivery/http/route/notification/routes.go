package notification

import (
	notifhandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/notification"
	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	h *notifhandler.Handler
}

func New(h *notifhandler.Handler) *Routes { return &Routes{h: h} }

func (rt *Routes) Register(r chi.Router, jwtSecret string) {
	r.Get("/ws/notify", rt.h.ServeWS)

	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Get("/member/notifications", rt.h.List)
		r.Put("/member/notifications/read-all", rt.h.MarkAllRead)
		r.Put("/member/notifications/{id}/read", rt.h.MarkRead)
	})
}

