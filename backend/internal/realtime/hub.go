package realtime

import (
	"context"

	"github.com/coder/websocket"
)

const clientSendBuffer = 64

// Client is a single WebSocket connection scoped to a userID. Only events for
// that userID are routed to its send channel.
type Client struct {
	userID string
	conn   *websocket.Conn
	send   chan []byte
}

// NewClient creates a Client bound to userID. userID must be non-empty; the
// handler is expected to reject empty ids before calling here.
func NewClient(userID string, conn *websocket.Conn) *Client {
	return &Client{
		userID: userID,
		conn:   conn,
		send:   make(chan []byte, clientSendBuffer),
	}
}

// UserID returns the user this client is scoped to.
func (c *Client) UserID() string { return c.userID }

// userBroadcast is an internal envelope sent on Hub.broadcast to fan out a
// payload only to clients whose userID matches.
type userBroadcast struct {
	userID  string
	payload []byte
}

type Hub struct {
	clients    map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan userBroadcast
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan userBroadcast, clientSendBuffer),
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

// BroadcastToUser fans the payload out to every connected client whose userID
// matches. Clients for other users never see this payload — preventing the
// pre-fix behavior where a single broadcast leaked to every connection.
func (h *Hub) BroadcastToUser(userID string, payload []byte) {
	if userID == "" || len(payload) == 0 {
		return
	}
	msg := append([]byte(nil), payload...)
	h.broadcast <- userBroadcast{userID: userID, payload: msg}
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
		case b := <-h.broadcast:
			for client := range h.clients {
				if client.userID != b.userID {
					continue
				}
				select {
				case client.send <- b.payload:
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
