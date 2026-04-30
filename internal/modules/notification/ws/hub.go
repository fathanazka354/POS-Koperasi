package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// NotifyHub mengelola koneksi WebSocket per member untuk push notifikasi.
type NotifyHub struct {
	mu      sync.RWMutex
	members map[int]*notifyClient
}

type notifyClient struct {
	conn *websocket.Conn
	send chan []byte
}

func NewNotifyHub() *NotifyHub {
	return &NotifyHub{members: make(map[int]*notifyClient)}
}

// Register mendaftarkan koneksi WS untuk member tertentu.
func (h *NotifyHub) Register(memberID int, conn *websocket.Conn) {
	h.mu.Lock()
	if old, ok := h.members[memberID]; ok {
		old.conn.Close()
	}
	c := &notifyClient{conn: conn, send: make(chan []byte, 16)}
	h.members[memberID] = c
	h.mu.Unlock()

	go h.writePump(memberID, c)
}

// Push mengirim payload JSON ke member tertentu jika sedang terkoneksi.
func (h *NotifyHub) Push(memberID int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.mu.RLock()
	c, ok := h.members[memberID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	select {
	case c.send <- data:
	default:
		log.Printf("notify hub: member %d send buffer full, skipping", memberID)
	}
}

func (h *NotifyHub) Unregister(memberID int) {
	h.mu.Lock()
	if c, ok := h.members[memberID]; ok {
		close(c.send)
		delete(h.members, memberID)
	}
	h.mu.Unlock()
}

func (h *NotifyHub) writePump(memberID int, c *notifyClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		h.Unregister(memberID)
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
