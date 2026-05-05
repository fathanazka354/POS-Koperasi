package notification

import (
	"net/http"

	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	notifyws "github.com/fathanazka354/pos-koperasi/internal/gateway/notifyws"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/notification"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
	uc     notification.Usecase
	hub    *notifyws.NotifyHub
	jwtKey string
}

func New(uc notification.Usecase, hub *notifyws.NotifyHub, jwtKey string) *Handler {
	return &Handler{uc: uc, hub: hub, jwtKey: jwtKey}
}

// GET /api/v1/member/notifications
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	items, err := h.uc.ListByMember(claims.MemberID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	unread, _ := h.uc.CountUnread(claims.MemberID)
	response.Success(w, "Notifikasi", map[string]interface{}{
		"items":  items,
		"unread": unread,
	})
}

// PUT /api/v1/member/notifications/{id}/read
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.uc.MarkRead(id, claims.MemberID); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Notifikasi ditandai dibaca", nil)
}

// PUT /api/v1/member/notifications/read-all
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	if err := h.uc.MarkAllRead(claims.MemberID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Semua notifikasi ditandai dibaca", nil)
}

// GET /api/v1/ws/notify?token=JWT
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := middleware.ParseMemberToken(tokenStr, h.jwtKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.Register(claims.MemberID, conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

