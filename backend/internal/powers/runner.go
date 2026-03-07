package powers

import (
	"context"

	"github.com/google/uuid"
)

type PlayerState struct {
	PlayerID       uuid.UUID
	Life           int
	Coins          int
	Gems           int
	HeroID         *string
	DeclaredHeroID *string
	IsWolfForm     bool
	IsEliminated   bool
}

type HeroInfo struct {
	ID       string
	Strength int
	Power1   string
	Power2   string
}

type MonsterInfo struct {
	ID         string
	Strength   int
	LootCoins  int
	LootGems   int
}

type DiscardEntry struct {
	ID     uuid.UUID
	HeroID string
}

type PowerRunner interface {
	GetGamePlayers(ctx context.Context, gameID uuid.UUID) ([]PlayerState, error)
	GetPlayerByID(ctx context.Context, gameID, playerID uuid.UUID) (*PlayerState, error)
	GetHeroByID(ctx context.Context, heroID string) (*HeroInfo, error)
	GetMonsterByID(ctx context.Context, monsterID string) (*MonsterInfo, error)
	GetActiveMonsterAtSlot(ctx context.Context, gameID uuid.UUID, slotIndex int) (monsterID string, err error)
	GetLastDiscardHeroID(ctx context.Context, gameID uuid.UUID) (*string, error)
	GetDiscardPile(ctx context.Context, gameID uuid.UUID) ([]DiscardEntry, error)

	AddPlayerLife(ctx context.Context, gameID, playerID uuid.UUID, delta int) error
	AddPlayerCoins(ctx context.Context, gameID, playerID uuid.UUID, delta int) error
	AddPlayerGems(ctx context.Context, gameID, playerID uuid.UUID, delta int) error
	UpdatePlayerHero(ctx context.Context, gameID, playerID uuid.UUID, heroID string) error
	ClearPlayerHero(ctx context.Context, gameID, playerID uuid.UUID) error
	AddHeroToDiscard(ctx context.Context, gameID uuid.UUID, heroID string, discardedBy *uuid.UUID) error
	DrawTopHeroFromDeck(ctx context.Context, gameID uuid.UUID) (heroID string, err error)
	ReplaceActiveMonster(ctx context.Context, gameID uuid.UUID, slotIndex int, monsterID string) error
	RemoveActiveMonsterSlot(ctx context.Context, gameID uuid.UUID, slotIndex int) error
	DrawRandomMonster(ctx context.Context) (monsterID string, err error)
	SetPlayerEliminated(ctx context.Context, gameID, playerID uuid.UUID, eliminated bool) error
	SetPlayerWolfForm(ctx context.Context, gameID, playerID uuid.UUID, isWolfForm bool) error

	ReturnDiscardToHeroDeckAndShuffle(ctx context.Context, gameID uuid.UUID) error
	BurnDiscardByID(ctx context.Context, gameID uuid.UUID, discardID uuid.UUID) error

	SetChallengerBonus(ctx context.Context, gameID uuid.UUID, playerID *uuid.UUID, bonus int) error
	GetChallengerBonus(ctx context.Context, gameID uuid.UUID) (playerID *uuid.UUID, bonus int, err error)
	SetDualAttackUsed(ctx context.Context, gameID uuid.UUID, used bool) error
	GetDualAttackUsed(ctx context.Context, gameID uuid.UUID) (bool, error)
	SetLastShootCorrect(ctx context.Context, gameID, playerID uuid.UUID, correct bool) error
	GetLastShootCorrect(ctx context.Context, gameID, playerID uuid.UUID) (bool, error)
	SetMimicHero(ctx context.Context, gameID uuid.UUID, actorID *uuid.UUID, heroID *string) error
	GetMimicHero(ctx context.Context, gameID uuid.UUID) (actorID *uuid.UUID, heroID *string, err error)

	EmitEvent(ctx context.Context, ev *Event)
	SecretReveal(ctx context.Context, gameID uuid.UUID, viewerID uuid.UUID, heroID string)
	ExecutePower(ctx context.Context, powerName string, ec ExecutionContext) error
}
