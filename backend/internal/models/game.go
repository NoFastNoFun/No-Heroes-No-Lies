package models

import (
	"time"

	"github.com/google/uuid"
)

type GameStatus string

const (
	GameStatusLobby    GameStatus = "lobby"
	GameStatusPlaying  GameStatus = "playing"
	GameStatusFinished GameStatus = "finished"
)

type Game struct {
	ID                 uuid.UUID
	CreatorID          uuid.UUID
	Status             GameStatus
	MaxPlayers         int
	CurrentTurnIndex   int
	RoundNumber        int
	WinConditionMet    bool
	WinnerID           *uuid.UUID
	LastDiscardHeroID  *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type GamePlayer struct {
	GameID          uuid.UUID
	PlayerID        uuid.UUID
	SlotIndex       int
	IsGameMaster    bool
	Life            int
	Coins           int
	Gems            int
	HeroID          *string
	DeclaredHeroID  *string
	IsWolfForm      bool
	IsEliminated    bool
	JoinedAt        time.Time
}

type GameListItem struct {
	ID          uuid.UUID
	CreatorID   uuid.UUID
	Status      GameStatus
	MaxPlayers  int
	PlayerCount int
	CreatedAt   time.Time
}
