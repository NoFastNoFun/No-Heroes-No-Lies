package ws

import (
	"net/http"
	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/services"

	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO: tighten for prod
}

// Message represents a client-to-server WebSocket message.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// GameWSHandler handles /ws/game/{id} WebSocket connections.
func GameWSHandler(w http.ResponseWriter, r *http.Request, sessionID string, gameSvc *services.GameService, sessionSvc *services.SessionService) {
	cookie, err := r.Cookie("game_auth")
	if err != nil || cookie.Value == "" {
		http.Error(w, "missing auth cookie", http.StatusUnauthorized)
		return
	}
	userID, _, err := auth.VerifyJWT(cookie.Value)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	// Membership check
	session, err := gameSvc.FetchSession(sessionID)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	isMember := false
	for _, id := range session.PlayerIDs {
		if id == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		for _, id := range session.State.SpectatorIDs {
			if id == userID {
				isMember = true
				break
			}
		}
	}
	if !isMember {
		http.Error(w, "not a member of this game", http.StatusForbidden)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{
		Conn:      conn,
		Send:      make(chan []byte, 256),
		UserID:    userID,
		SessionID: sessionID,
	}
	hub := GetHub(sessionID)
	hub.Register <- client
	defer func() { hub.Unregister <- client; conn.Close() }()
	// Start write pump
	go func() {
		for msg := range client.Send {
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()
	// Read loop: handle client messages
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("ws: invalid message: %v", err)
			continue
		}
		switch msg.Type {
		case "move":
			var movePayload models.MovePayload
			if err := json.Unmarshal(msg.Payload, &movePayload); err == nil {
				err := gameSvc.ApplyMove(sessionID, userID, movePayload)
				if err == nil {
					updated, err := gameSvc.FetchSession(sessionID)
					if err == nil {
						stateJSON, _ := json.Marshal(updated)
						hub.Broadcast <- stateJSON
					}
				}
			}
		case "ready":
			_, err := sessionSvc.ToggleReady(sessionID, userID)
			if err == nil {
				updated, err := gameSvc.FetchSession(sessionID)
				if err == nil {
					stateJSON, _ := json.Marshal(updated)
					hub.Broadcast <- stateJSON
				}
			}
		case "forfeit":
			err := gameSvc.Forfeit(sessionID, userID)
			if err == nil {
				updated, err := gameSvc.FetchSession(sessionID)
				if err == nil {
					stateJSON, _ := json.Marshal(updated)
					hub.Broadcast <- stateJSON
				}
			}
		}
	}
}
