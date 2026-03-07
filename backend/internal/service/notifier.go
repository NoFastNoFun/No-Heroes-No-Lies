package service

import (
	"time"

	"github.com/google/uuid"
)

type GameNotifier interface {
	NotifyStateUpdated(gameID uuid.UUID)
	NotifyChallengeWindowOpen(gameID uuid.UUID, endsAt time.Time)
	NotifyChallengeWindowClosed(gameID uuid.UUID)
	NotifySecretReveal(gameID uuid.UUID, viewerID uuid.UUID, heroID string)
}
