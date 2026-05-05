package chat

import (
	chathandler "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/chat"
	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Routes struct {
	h *chathandler.Handler
}

func New(h *chathandler.Handler) *Routes { return &Routes{h: h} }

func (rt *Routes) Register(r chi.Router, jwtSecret string) {
	r.Get("/ws/chat", rt.h.ServeWS)

	r.Get("/chat/conversations", rt.h.ListConversations)
	r.Get("/chat/conversations/{id}/presence", rt.h.ConversationPresence)
	r.Get("/chat/conversations/{id}/messages", rt.h.ListMessages)

	r.Group(func(r chi.Router) {
		r.Use(middleware.MemberJWTAuth(jwtSecret))
		r.Post("/chat/conversations", rt.h.CreateConversation)
	})
}

