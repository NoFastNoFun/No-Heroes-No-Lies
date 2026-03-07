package models

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID        uuid.UUID
	Username  string
	CreatedAt time.Time
}

type Session struct {
	ID        uuid.UUID
	PlayerID  uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
