package ws

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/service"
)

type Notifier struct {
	hub *Hub
}

func NewNotifier(hub *Hub) service.GameNotifier {
	return &Notifier{hub: hub}
}

func (n *Notifier) NotifyStateUpdated(gameID uuid.UUID) {
	msg, _ := json.Marshal(map[string]string{"type": "state_updated"})
	n.hub.BroadcastToGame(gameID, msg)
}

func (n *Notifier) NotifyChallengeWindowOpen(gameID uuid.UUID, endsAt time.Time) {
	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "challenge_window_open",
		"ends_at": endsAt.Format(time.RFC3339),
	})
	n.hub.BroadcastToGame(gameID, msg)
}

func (n *Notifier) NotifyChallengeWindowClosed(gameID uuid.UUID) {
	msg, _ := json.Marshal(map[string]string{"type": "challenge_window_closed"})
	n.hub.BroadcastToGame(gameID, msg)
}

func (n *Notifier) NotifySecretReveal(gameID uuid.UUID, viewerID uuid.UUID, heroID string) {
	msg, _ := json.Marshal(map[string]string{
		"type":    "secret_reveal",
		"hero_id": heroID,
	})
	n.hub.SendToPlayer(gameID, viewerID, msg)
}
