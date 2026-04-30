package controller

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/notification/contract"
	notifyws "github.com/yourname/pos-koperasi/internal/modules/notification/ws"
	"github.com/yourname/pos-koperasi/pkg/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Controller struct {
	svc    contract.NotificationService
	hub    *notifyws.NotifyHub
	jwtKey string
}

func New(svc contract.NotificationService, hub *notifyws.NotifyHub, jwtKey string) *Controller {
	return &Controller{svc: svc, hub: hub, jwtKey: jwtKey}
}

// GET /api/v1/member/notifications
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	items, err := c.svc.ListByMember(claims.MemberID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	unread, _ := c.svc.CountUnread(claims.MemberID)
	response.Success(w, "Notifikasi", map[string]interface{}{
		"items":  items,
		"unread": unread,
	})
}

// PUT /api/v1/member/notifications/{id}/read
func (c *Controller) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	id := chi.URLParam(r, "id")
	if err := c.svc.MarkRead(id, claims.MemberID); err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "Notifikasi ditandai dibaca", nil)
}

// PUT /api/v1/member/notifications/read-all
func (c *Controller) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}
	if err := c.svc.MarkAllRead(claims.MemberID); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Semua notifikasi ditandai dibaca", nil)
}

// GET /api/v1/ws/notify?token=JWT  — member WebSocket untuk push notifikasi
func (c *Controller) ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	claims, err := middleware.ParseMemberToken(tokenStr, c.jwtKey)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	c.hub.Register(claims.MemberID, conn)

	// Baca frame masuk supaya koneksi tetap hidup (kita abaikan isinya)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
