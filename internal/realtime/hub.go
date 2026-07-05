package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[*Client]bool)}
}

func (h *Hub) Register(topic string, conn *websocket.Conn) *Client {
	c := &Client{conn: conn, send: make(chan []byte, 16)}

	h.mu.Lock()
	if h.clients[topic] == nil {
		h.clients[topic] = make(map[*Client]bool)
	}
	h.clients[topic][c] = true
	h.mu.Unlock()

	go c.writePump()
	return c
}

func (h *Hub) Unregister(topic string, c *Client) {
	h.mu.Lock()
	if clients, ok := h.clients[topic]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(h.clients, topic)
		}
	}
	h.mu.Unlock()
	close(c.send)
}

func (h *Hub) Broadcast(topic string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[topic] {
		select {
		case c.send <- message:
		default:
		}
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *Client) ReadLoop() {
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}
