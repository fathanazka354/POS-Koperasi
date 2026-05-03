package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/modules/chat/contract"
	"github.com/yourname/pos-koperasi/internal/modules/chat/controller/dto"
	"github.com/yourname/pos-koperasi/internal/modules/chat/utility"
	"github.com/yourname/pos-koperasi/pkg/response"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Controller struct {
	svc       contract.ChatService
	jwtSecret string
}

func New(svc contract.ChatService, jwtSecret string) *Controller {
	return &Controller{svc: svc, jwtSecret: jwtSecret}
}

// POST /api/v1/chat/conversations (member JWT)
func (c *Controller) CreateConversation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req dto.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	conv, err := c.svc.CreateConversation(claims.MemberID, contract.CreateConversationInput{
		ProductID:        req.ProductID,
		SellerEmployeeID: req.SellerEmployeeID,
		FirstMessage:     req.FirstMessage,
	})
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, "Percakapan dibuat", conv)
}

// GET /api/v1/chat/conversations — token karyawan atau member
func (c *Controller) ListConversations(w http.ResponseWriter, r *http.Request) {
	p, err := c.chatPrincipalFromRequest(r)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	list, err := c.svc.ListConversations(p)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, "OK", list)
}

// GET /api/v1/chat/conversations/{id}/messages
func (c *Controller) ListMessages(w http.ResponseWriter, r *http.Request) {
	p, err := c.chatPrincipalFromRequest(r)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	idStr := chi.URLParam(r, "id")
	convID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || convID <= 0 {
		response.BadRequest(w, "id tidak valid")
		return
	}

	msgs, err := c.svc.ListMessages(convID, p)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, "OK", msgs)
}

// GET /api/v1/chat/conversations/{id}/presence
func (c *Controller) ConversationPresence(w http.ResponseWriter, r *http.Request) {
	p, err := c.chatPrincipalFromRequest(r)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}
	idStr := chi.URLParam(r, "id")
	convID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || convID <= 0 {
		response.BadRequest(w, "id tidak valid")
		return
	}
	pr, err := c.svc.ConversationPresence(convID, p)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "OK", map[string]bool{
		"buyer_online":  pr.BuyerOnline,
		"seller_online": pr.SellerOnline,
	})
}

func (c *Controller) chatPrincipalFromRequest(r *http.Request) (*middleware.ChatPrincipal, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, fmt.Errorf("authorization required")
	}
	const prefix = "Bearer "
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		return nil, fmt.Errorf("invalid authorization format")
	}
	token := auth[len(prefix):]
	return middleware.ParseChatToken(token, c.jwtSecret)
}

// ServeWS GET /api/v1/ws/chat?token=JWT — karyawan atau member
func (c *Controller) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token query required", http.StatusUnauthorized)
		return
	}

	principal, err := middleware.ParseChatToken(token, c.jwtSecret)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}

	client := utility.NewClient(c.svc.Hub(), conn, 256)
	c.svc.PresenceConnect(principal)
	defer func() {
		c.svc.PresenceDisconnect(principal)
	}()
	defer close(client.Send)
	go c.wsWritePump(client)
	c.wsReadPump(client, principal)
}

func (c *Controller) wsWritePump(cl *utility.Client) {
	defer func() {
		c.svc.Hub().LeaveAllRooms(cl)
		_ = cl.Conn.Close()
	}()
	for msg := range cl.Send {
		if err := cl.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Controller) wsReadPump(cl *utility.Client, p *middleware.ChatPrincipal) {
	defer func() {
		c.svc.Hub().LeaveAllRooms(cl)
	}()

	for {
		_, data, err := cl.Conn.ReadMessage()
		if err != nil {
			return
		}

		var msg dto.WSClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "join":
			if msg.ConversationID <= 0 {
				c.wsSendError(cl, "conversation_id wajib untuk join")
				continue
			}
			if err := c.svc.JoinRoom(msg.ConversationID, p, cl); err != nil {
				c.wsSendError(cl, err.Error())
				continue
			}
			ok, _ := json.Marshal(dto.WSServerMessage{Type: "joined", ConversationID: msg.ConversationID})
			select {
			case cl.Send <- ok:
			default:
			}

		case "message":
			if msg.ConversationID <= 0 {
				c.wsSendError(cl, "conversation_id wajib")
				continue
			}
			// Langkah join implisit: yang kirim pesan otomatis masuk room, supaya broadcast
			// pesan lawan sampai tanpa harus klik "Join room" dulu (Join eksplisit tetap didukung).
			if err := c.svc.JoinRoom(msg.ConversationID, p, cl); err != nil {
				c.wsSendError(cl, err.Error())
				continue
			}
			payload, err := c.svc.PostMessage(msg.ConversationID, msg.Body, p)
			if err != nil {
				c.wsSendError(cl, err.Error())
				continue
			}
			c.svc.Hub().BroadcastRoom(msg.ConversationID, payload)

		case "ping":
			ok, _ := json.Marshal(dto.WSServerMessage{Type: "pong"})
			select {
			case cl.Send <- ok:
			default:
			}

		default:
			c.wsSendError(cl, "type tidak dikenal (gunakan join, message, atau ping)")
		}
	}
}

func (c *Controller) wsSendError(cl *utility.Client, errText string) {
	b, _ := json.Marshal(dto.WSServerMessage{Type: "error", Error: errText})
	select {
	case cl.Send <- b:
	default:
	}
}

