package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/no-heroes-no-lies/backend/internal/gameloop"
	"github.com/no-heroes-no-lies/backend/internal/models"
	"github.com/no-heroes-no-lies/backend/internal/repository"
	"github.com/no-heroes-no-lies/backend/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu      sync.RWMutex
	rooms   map[uuid.UUID]map[*Client]struct{}
	registry *gameloop.Registry
	session *service.SessionService
	db      *repository.DB
}

func NewHub(registry *gameloop.Registry, session *service.SessionService, db *repository.DB) *Hub {
	return &Hub{
		rooms:    make(map[uuid.UUID]map[*Client]struct{}),
		registry: registry,
		session:  session,
		db:       db,
	}
}

func (h *Hub) SetRegistry(registry *gameloop.Registry) {
	h.mu.Lock()
	h.registry = registry
	h.mu.Unlock()
}

type Client struct {
	conn     *websocket.Conn
	playerID uuid.UUID
	gameID   uuid.UUID
	send     chan []byte
	hub      *Hub
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	if h.rooms[c.gameID] == nil {
		h.rooms[c.gameID] = make(map[*Client]struct{})
	}
	h.rooms[c.gameID][c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if room, ok := h.rooms[c.gameID]; ok {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, c.gameID)
		}
	}
	h.mu.Unlock()
	close(c.send)
}

func (h *Hub) BroadcastToGame(gameID uuid.UUID, msg []byte) {
	h.mu.RLock()
	room := h.rooms[gameID]
	for c := range room {
		select {
		case c.send <- msg:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) SendToPlayer(gameID, playerID uuid.UUID, msg []byte) {
	h.mu.RLock()
	room := h.rooms[gameID]
	for c := range room {
		if c.playerID == playerID {
			select {
			case c.send <- msg:
			default:
			}
			break
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request, gameID uuid.UUID) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}
	playerID, err := h.session.ValidateToken(r.Context(), token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	inGame, err := h.db.IsPlayerInGame(r.Context(), gameID, playerID)
	if err != nil || !inGame {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	game, err := h.db.GetGameByID(r.Context(), gameID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if game.Status != models.GameStatusPlaying {
		http.Error(w, "game not in play", http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{
		conn:     conn,
		playerID: playerID,
		gameID:   gameID,
		send:     make(chan []byte, 256),
		hub:      h,
	}
	h.Register(client)
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var envelope struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(message, &envelope); err != nil {
			continue
		}
		cmd := gameloop.Command{
			Type:     gameloop.CommandType(envelope.Type),
			PlayerID: c.playerID,
			Payload:  envelope.Payload,
		}
		if cmd.Type == "" {
			continue
		}
		if c.hub.registry != nil {
			ok := c.hub.registry.Send(context.Background(), c.gameID, cmd)
			if !ok {
				slog.Debug("game loop not running", "game_id", c.gameID)
			}
		}
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
