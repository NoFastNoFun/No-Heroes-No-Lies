package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/powers"
	"github.com/no-heroes-no-lies/backend/internal/repository"
)

type PowerRunnerImpl struct {
	db       *repository.DB
	notifier GameNotifier
	registry *powers.Registry
}

func NewPowerRunner(db *repository.DB, notifier GameNotifier, registry *powers.Registry) *PowerRunnerImpl {
	return &PowerRunnerImpl{db: db, notifier: notifier, registry: registry}
}

func (r *PowerRunnerImpl) GetGamePlayers(ctx context.Context, gameID uuid.UUID) ([]powers.PlayerState, error) {
	list, err := r.db.GetGamePlayers(ctx, gameID)
	if err != nil {
		return nil, err
	}
	out := make([]powers.PlayerState, len(list))
	for i := range list {
		out[i] = powers.PlayerState{
			PlayerID:       list[i].PlayerID,
			Life:           list[i].Life,
			Coins:          list[i].Coins,
			Gems:           list[i].Gems,
			HeroID:         list[i].HeroID,
			DeclaredHeroID: list[i].DeclaredHeroID,
			IsWolfForm:     list[i].IsWolfForm,
			IsEliminated:   list[i].IsEliminated,
		}
	}
	return out, nil
}

func (r *PowerRunnerImpl) GetPlayerByID(ctx context.Context, gameID, playerID uuid.UUID) (*powers.PlayerState, error) {
	gp, err := r.db.GetPlayerByID(ctx, gameID, playerID)
	if err != nil {
		return nil, err
	}
	return &powers.PlayerState{
		PlayerID:       gp.PlayerID,
		Life:           gp.Life,
		Coins:          gp.Coins,
		Gems:           gp.Gems,
		HeroID:         gp.HeroID,
		DeclaredHeroID: gp.DeclaredHeroID,
		IsWolfForm:     gp.IsWolfForm,
		IsEliminated:   gp.IsEliminated,
	}, nil
}

func (r *PowerRunnerImpl) GetHeroByID(ctx context.Context, heroID string) (*powers.HeroInfo, error) {
	h, err := r.db.GetHeroByID(ctx, heroID)
	if err != nil {
		return nil, err
	}
	return &powers.HeroInfo{ID: h.ID, Strength: h.Strength, Power1: h.Power1, Power2: h.Power2}, nil
}

func (r *PowerRunnerImpl) GetMonsterByID(ctx context.Context, monsterID string) (*powers.MonsterInfo, error) {
	m, err := r.db.GetMonsterByID(ctx, monsterID)
	if err != nil {
		return nil, err
	}
	return &powers.MonsterInfo{ID: m.ID, Strength: m.Strength, LootCoins: m.LootCoins, LootGems: m.LootGems}, nil
}

func (r *PowerRunnerImpl) GetActiveMonsterAtSlot(ctx context.Context, gameID uuid.UUID, slotIndex int) (string, error) {
	return r.db.GetActiveMonsterAtSlot(ctx, gameID, slotIndex)
}

func (r *PowerRunnerImpl) GetLastDiscardHeroID(ctx context.Context, gameID uuid.UUID) (*string, error) {
	g, err := r.db.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	return g.LastDiscardHeroID, nil
}

func (r *PowerRunnerImpl) GetDiscardPile(ctx context.Context, gameID uuid.UUID) ([]powers.DiscardEntry, error) {
	rows, err := r.db.GetDiscardPile(ctx, gameID)
	if err != nil {
		return nil, err
	}
	out := make([]powers.DiscardEntry, len(rows))
	for i := range rows {
		out[i] = powers.DiscardEntry{ID: rows[i].ID, HeroID: rows[i].HeroID}
	}
	return out, nil
}

func (r *PowerRunnerImpl) AddPlayerLife(ctx context.Context, gameID, playerID uuid.UUID, delta int) error {
	return r.db.AddPlayerLife(ctx, gameID, playerID, delta)
}

func (r *PowerRunnerImpl) AddPlayerCoins(ctx context.Context, gameID, playerID uuid.UUID, delta int) error {
	return r.db.AddPlayerCoins(ctx, gameID, playerID, delta)
}

func (r *PowerRunnerImpl) AddPlayerGems(ctx context.Context, gameID, playerID uuid.UUID, delta int) error {
	return r.db.AddPlayerGems(ctx, gameID, playerID, delta)
}

func (r *PowerRunnerImpl) UpdatePlayerHero(ctx context.Context, gameID, playerID uuid.UUID, heroID string) error {
	return r.db.UpdatePlayerHero(ctx, gameID, playerID, heroID)
}

func (r *PowerRunnerImpl) ClearPlayerHero(ctx context.Context, gameID, playerID uuid.UUID) error {
	return r.db.ClearPlayerHero(ctx, gameID, playerID)
}

func (r *PowerRunnerImpl) AddHeroToDiscard(ctx context.Context, gameID uuid.UUID, heroID string, discardedBy *uuid.UUID) error {
	return r.db.AddHeroToDiscard(ctx, gameID, heroID, discardedBy)
}

func (r *PowerRunnerImpl) DrawTopHeroFromDeck(ctx context.Context, gameID uuid.UUID) (string, error) {
	return r.db.DrawTopHeroFromDeck(ctx, gameID)
}

func (r *PowerRunnerImpl) ReplaceActiveMonster(ctx context.Context, gameID uuid.UUID, slotIndex int, monsterID string) error {
	return r.db.ReplaceActiveMonster(ctx, gameID, slotIndex, monsterID)
}

func (r *PowerRunnerImpl) RemoveActiveMonsterSlot(ctx context.Context, gameID uuid.UUID, slotIndex int) error {
	return r.db.RemoveActiveMonsterSlot(ctx, gameID, slotIndex)
}

func (r *PowerRunnerImpl) DrawRandomMonster(ctx context.Context) (string, error) {
	return r.db.DrawRandomMonsterByWeight(ctx)
}

func (r *PowerRunnerImpl) SetPlayerEliminated(ctx context.Context, gameID, playerID uuid.UUID, eliminated bool) error {
	return r.db.SetPlayerEliminated(ctx, gameID, playerID, eliminated)
}

func (r *PowerRunnerImpl) SetPlayerWolfForm(ctx context.Context, gameID, playerID uuid.UUID, isWolfForm bool) error {
	return r.db.UpdatePlayerWolfForm(ctx, gameID, playerID, isWolfForm)
}

func (r *PowerRunnerImpl) ReturnDiscardToHeroDeckAndShuffle(ctx context.Context, gameID uuid.UUID) error {
	return r.db.ReturnDiscardToHeroDeckAndShuffle(ctx, gameID)
}

func (r *PowerRunnerImpl) BurnDiscardByID(ctx context.Context, gameID uuid.UUID, discardID uuid.UUID) error {
	return r.db.BurnDiscardByID(ctx, gameID, discardID)
}

func (r *PowerRunnerImpl) SetChallengerBonus(ctx context.Context, gameID uuid.UUID, playerID *uuid.UUID, bonus int) error {
	return r.db.SetChallengerBonus(ctx, gameID, playerID, bonus)
}

func (r *PowerRunnerImpl) GetChallengerBonus(ctx context.Context, gameID uuid.UUID) (*uuid.UUID, int, error) {
	return r.db.GetChallengerBonus(ctx, gameID)
}

func (r *PowerRunnerImpl) SetDualAttackUsed(ctx context.Context, gameID uuid.UUID, used bool) error {
	return r.db.SetDualAttackUsed(ctx, gameID, used)
}

func (r *PowerRunnerImpl) GetDualAttackUsed(ctx context.Context, gameID uuid.UUID) (bool, error) {
	return r.db.GetDualAttackUsed(ctx, gameID)
}

func (r *PowerRunnerImpl) SetLastShootCorrect(ctx context.Context, gameID, playerID uuid.UUID, correct bool) error {
	return r.db.SetLastShootCorrect(ctx, gameID, playerID, correct)
}

func (r *PowerRunnerImpl) GetLastShootCorrect(ctx context.Context, gameID, playerID uuid.UUID) (bool, error) {
	return r.db.GetLastShootCorrect(ctx, gameID, playerID)
}

func (r *PowerRunnerImpl) SetMimicHero(ctx context.Context, gameID uuid.UUID, actorID *uuid.UUID, heroID *string) error {
	return r.db.SetMimicHero(ctx, gameID, actorID, heroID)
}

func (r *PowerRunnerImpl) GetMimicHero(ctx context.Context, gameID uuid.UUID) (*uuid.UUID, *string, error) {
	return r.db.GetMimicHero(ctx, gameID)
}

func (r *PowerRunnerImpl) EmitEvent(ctx context.Context, ev *powers.Event) {
	ev.Runner = r
	powers.EmitEvent(ctx, ev)
}

func (r *PowerRunnerImpl) SecretReveal(ctx context.Context, gameID uuid.UUID, viewerID uuid.UUID, heroID string) {
	if r.notifier != nil {
		r.notifier.NotifySecretReveal(gameID, viewerID, heroID)
	}
}

func (r *PowerRunnerImpl) ExecutePower(ctx context.Context, powerName string, ec powers.ExecutionContext) error {
	if r.registry == nil {
		return nil
	}
	return r.registry.Execute(ctx, powerName, ec, r)
}
