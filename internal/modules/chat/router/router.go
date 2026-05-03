package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/chat/controller"
)

type Router struct {
	chat *controller.Controller
}

func New(chat *controller.Controller) *Router {
	return &Router{chat: chat}
}

func (rt *Router) Register(r chi.Router, jwtSecret string) {
	// WebSocket chat (token lewat query ?token= untuk browser)
	r.Get("/ws/chat", rt.chat.ServeWS)

	// Chat — daftar & riwayat (Bearer karyawan atau member; validasi di controller)
	r.Get("/chat/conversations", rt.chat.ListConversations)
	r.Get("/chat/conversations/{id}/presence", rt.chat.ConversationPresence)
	r.Get("/chat/conversations/{id}/messages", rt.chat.ListMessages)

	// Buat percakapan — hanya member (JWT member)
	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Post("/chat/conversations", rt.chat.CreateConversation)
	})
}

