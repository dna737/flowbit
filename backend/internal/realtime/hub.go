package realtime

import (
	"context"

	"github.com/coder/websocket"
)

const clientSendBuffer = 64

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, clientSendBuffer),
	}
}

type Hub struct {
	clients    map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, clientSendBuffer),
	}
}

func (h *Hub) Register(client *Client) {
	if client == nil {
		return
	}
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}
	h.unregister <- client
}

func (h *Hub) Broadcast(payload []byte) {
	if len(payload) == 0 {
		return
	}
	msg := append([]byte(nil), payload...)
	h.broadcast <- msg
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for client := range h.clients {
				h.removeClient(client)
			}
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			h.removeClient(client)
		case payload := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- payload:
				default:
					h.removeClient(client)
				}
			}
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	if _, ok := h.clients[client]; !ok {
		return
	}
	delete(h.clients, client)
	close(client.send)
}
