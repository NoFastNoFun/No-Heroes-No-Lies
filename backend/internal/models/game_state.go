package models

import (
	"time"

	"github.com/google/uuid"
)

type TurnPhase string

const (
	PhasePick             TurnPhase = "pick"
	PhaseDiscard          TurnPhase = "discard"
	PhaseDeclare          TurnPhase = "declare"
	PhasePower            TurnPhase = "power"
	PhaseAttack           TurnPhase = "attack"
	PhaseDemask           TurnPhase = "demask"
	PhaseResolvingPassives TurnPhase = "resolving_passives"
	PhaseBetweenTurns     TurnPhase = "between_turns"
)

type GameTurnState struct {
	GameID                 uuid.UUID
	Phase                  TurnPhase
	PhaseEnteredAt         time.Time
	ChallengeWindowEndsAt  *time.Time
	PendingActionType      *string
	PendingActionPayload   []byte
	ActingPlayerID         *uuid.UUID
	PowerStep              *string
	PickedHeroID           *string
}

type GameStateView struct {
	GameID        uuid.UUID
	Status        GameStatus
	MaxPlayers    int
	RoundNumber   int
	LastDiscardID *string
	Phase         TurnPhase
	ActingPlayerID *uuid.UUID
	ChallengeWindowEndsAt *time.Time
	Players       []PlayerStateView
	ActiveMonsters []ActiveMonsterView
	MySlot        int
	MyHeroID      *string
	MyLife        int
	MyCoins       int
	MyGems        int
	DeckSize      int
	WinnerID      *uuid.UUID
}

type PlayerStateView struct {
	PlayerID   uuid.UUID
	SlotIndex  int
	Life       int
	Coins      int
	Gems       int
	HeroID     *string
	DeclaredHeroID *string
	IsEliminated bool
}

type ActiveMonsterView struct {
	SlotIndex int
	MonsterID string
}
