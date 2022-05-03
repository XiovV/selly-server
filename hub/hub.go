package hub

import (
	"github.com/gorilla/websocket"
)

type userKey string

type Client struct {
	Username string
	Conn     *websocket.Conn
}

type Hub struct {
	connections map[userKey]*websocket.Conn
}

func New() *Hub {
	return &Hub{connections: make(map[userKey]*websocket.Conn)}
}

func (h *Hub) Get(username string) (*websocket.Conn, bool) {
	conn, exists := h.connections[userKey(username)]

	return conn, exists
}

func (h *Hub) Push(client Client) {
	h.connections[userKey(client.Username)] = client.Conn
}

func (h *Hub) Pop(username string) {
	delete(h.connections, userKey(username))
}
