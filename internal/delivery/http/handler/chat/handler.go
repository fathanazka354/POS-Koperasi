package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	chatdto "github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/chat/dto"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/chatws/utility"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/chat"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	uc        chat.Usecase
	jwtSecret string
}

func New(uc chat.Usecase, jwtSecret string) *Handler {
	return &Handler{uc: uc, jwtSecret: jwtSecret}
}

// POST /api/v1/chat/conversations (member JWT)
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetMemberClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req chatdto.CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	conv, err := h.uc.CreateConversation(claims.MemberID, chat.CreateConversationInput{
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
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	p, err := h.chatPrincipalFromRequest(r)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	list, err := h.uc.ListConversations(p)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Success(w, "OK", list)
}

// GET /api/v1/chat/conversations/{id}/messages
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	p, err := h.chatPrincipalFromRequest(r)
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

	msgs, err := h.uc.ListMessages(convID, p)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, "OK", msgs)
}

// GET /api/v1/chat/conversations/{id}/presence
func (h *Handler) ConversationPresence(w http.ResponseWriter, r *http.Request) {
	p, err := h.chatPrincipalFromRequest(r)
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
	pr, err := h.uc.ConversationPresence(convID, p)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}
	response.Success(w, "OK", map[string]bool{
		"buyer_online":  pr.BuyerOnline,
		"seller_online": pr.SellerOnline,
	})
}

func (h *Handler) chatPrincipalFromRequest(r *http.Request) (*chat.ChatPrincipal, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, fmt.Errorf("authorization required")
	}
	const prefix = "Bearer "
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		return nil, fmt.Errorf("invalid authorization format")
	}
	token := auth[len(prefix):]
	return middleware.ParseChatToken(token, h.jwtSecret)
}

// ServeWS GET /api/v1/ws/chat?token=JWT — karyawan atau member
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token query required", http.StatusUnauthorized)
		return
	}

	principal, err := middleware.ParseChatToken(token, h.jwtSecret)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}

	client := utility.NewClient(h.uc.Hub(), conn, 256)
	h.uc.PresenceConnect(principal)
	defer func() {
		h.uc.PresenceDisconnect(principal)
	}()
	defer close(client.Send)
	go h.wsWritePump(client)
	h.wsReadPump(client, principal)
}

func (h *Handler) wsWritePump(cl *utility.Client) {
	defer func() {
		h.uc.Hub().LeaveAllRooms(cl)
		_ = cl.Conn.Close()
	}()
	for msg := range cl.Send {
		if err := cl.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (h *Handler) wsReadPump(cl *utility.Client, p *chat.ChatPrincipal) {
	defer func() {
		h.uc.Hub().LeaveAllRooms(cl)
	}()

	for {
		_, data, err := cl.Conn.ReadMessage()
		if err != nil {
			return
		}

		var msg chatdto.WSClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "join":
			if msg.ConversationID <= 0 {
				h.wsSendError(cl, "conversation_id wajib untuk join")
				continue
			}
			if err := h.uc.JoinRoom(msg.ConversationID, p, cl); err != nil {
				h.wsSendError(cl, err.Error())
				continue
			}
			ok, _ := json.Marshal(chatdto.WSServerMessage{Type: "joined", ConversationID: msg.ConversationID})
			select {
			case cl.Send <- ok:
			default:
			}

		case "message":
			if msg.ConversationID <= 0 {
				h.wsSendError(cl, "conversation_id wajib")
				continue
			}
			if err := h.uc.JoinRoom(msg.ConversationID, p, cl); err != nil {
				h.wsSendError(cl, err.Error())
				continue
			}
			payload, err := h.uc.PostMessage(msg.ConversationID, msg.Body, p)
			if err != nil {
				h.wsSendError(cl, err.Error())
				continue
			}
			h.uc.Hub().BroadcastRoom(msg.ConversationID, payload)

		case "ping":
			ok, _ := json.Marshal(chatdto.WSServerMessage{Type: "pong"})
			select {
			case cl.Send <- ok:
			default:
			}

		default:
			h.wsSendError(cl, "type tidak dikenal (gunakan join, message, atau ping)")
		}
	}
}

func (h *Handler) wsSendError(cl *utility.Client, errText string) {
	b, _ := json.Marshal(chatdto.WSServerMessage{Type: "error", Error: errText})
	select {
	case cl.Send <- b:
	default:
	}
}

