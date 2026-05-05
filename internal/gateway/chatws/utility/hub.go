package utility

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub mengelola ruang percakapan realtime (conversation_id -> koneksi).
type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[int64]map[*Client]struct{})}
}

func (h *Hub) JoinRoom(convID int64, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[convID] == nil {
		h.rooms[convID] = make(map[*Client]struct{})
	}
	h.rooms[convID][c] = struct{}{}
	c.mu.Lock()
	c.rooms[convID] = struct{}{}
	c.mu.Unlock()
}

func (h *Hub) LeaveAllRooms(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c.mu.Lock()
	for convID := range c.rooms {
		if set, ok := h.rooms[convID]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(h.rooms, convID)
			}
		}
	}
	c.rooms = make(map[int64]struct{})
	c.mu.Unlock()
}

// BroadcastRoom mengirim payload ke semua klien yang sudah join room (termasuk pengirim).
func (h *Hub) BroadcastRoom(convID int64, payload []byte) {
	h.mu.RLock()
	set := h.rooms[convID]
	clients := make([]*Client, 0, len(set))
	for cl := range set {
		clients = append(clients, cl)
	}
	h.mu.RUnlock()

	for _, cl := range clients {
		select {
		case cl.Send <- payload:
		default:
			go cl.CloseSlow()
		}
	}
}

// Client satu koneksi WebSocket; bisa join beberapa conversation_id.
type Client struct {
	hub   *Hub
	Conn  *websocket.Conn
	Send  chan []byte
	mu    sync.Mutex
	rooms map[int64]struct{}
}

func NewClient(hub *Hub, conn *websocket.Conn, sendBuf int) *Client {
	if sendBuf <= 0 {
		sendBuf = 256
	}
	return &Client{
		hub:   hub,
		Conn:  conn,
		Send:  make(chan []byte, sendBuf),
		rooms: make(map[int64]struct{}),
	}
}

func (c *Client) CloseSlow() {
	_ = c.Conn.Close()
}

