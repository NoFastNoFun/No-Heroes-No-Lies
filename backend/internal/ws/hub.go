package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	Send      chan []byte
	UserID    string
	SessionID string
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

var sessionHubs = struct {
	m map[string]*Hub
	sync.RWMutex
}{m: make(map[string]*Hub)}

// GetHub returns the hub for a session, creating it if needed.
func GetHub(sessionID string) *Hub {
	sessionHubs.Lock()
	hub, ok := sessionHubs.m[sessionID]
	if !ok {
		hub = &Hub{
			Clients:    make(map[*Client]bool),
			Broadcast:  make(chan []byte),
			Register:   make(chan *Client),
			Unregister: make(chan *Client),
		}
		sessionHubs.m[sessionID] = hub
		go hub.run()
	}
	sessionHubs.Unlock()
	return hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}
