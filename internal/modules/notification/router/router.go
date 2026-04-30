package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/notification/controller"
)

type Router struct {
	ctl *controller.Controller
}

func New(ctl *controller.Controller) *Router {
	return &Router{ctl: ctl}
}

func (rt *Router) Register(r chi.Router, jwtSecret string) {
	// Public WebSocket endpoint (token via query)
	r.Get("/ws/notify", rt.ctl.ServeWS)

	// Protected member routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Get("/member/notifications", rt.ctl.List)
		r.Put("/member/notifications/read-all", rt.ctl.MarkAllRead)
		r.Put("/member/notifications/{id}/read", rt.ctl.MarkRead)
	})
}
