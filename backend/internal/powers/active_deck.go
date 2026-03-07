package powers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

func RegisterActiveDeck(reg *Registry) {
	reg.Register("change_hero", PowerDef{Name: "change_hero", CostGems: 2, TargetType: TargetSelf}, execChangeHero)
	reg.Register("shuffle_heroes_deck", PowerDef{Name: "shuffle_heroes_deck", CostGems: 0, TargetType: TargetSession}, execShuffleHeroesDeck)
	reg.Register("change_monster", PowerDef{Name: "change_monster", CostGems: 0, TargetType: TargetMonster}, execChangeMonster)
	reg.Register("blind_draw", PowerDef{Name: "blind_draw", CostGems: 0, TargetType: TargetOtherPlayer}, execBlindDraw)
	reg.Register("force_transform", PowerDef{Name: "force_transform", CostGems: 4, TargetType: TargetOtherPlayer}, execForceTransform)
	reg.Register("execution", PowerDef{Name: "execution", CostGems: 6, TargetType: TargetSession}, execExecution)
	reg.Register("command_the_dead", PowerDef{Name: "command_the_dead", CostGems: 4, TargetType: TargetMonster}, execCommandTheDead)
}

func execChangeHero(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	actor, err := r.GetPlayerByID(ctx, ec.GameID, ec.ActorID)
	if err != nil {
		return err
	}
	if actor.HeroID == nil {
		return errBadRequest
	}
	currentHero := *actor.HeroID
	newHeroID, err := r.DrawTopHeroFromDeck(ctx, ec.GameID)
	if err != nil {
		return err
	}
	if err := r.AddHeroToDiscard(ctx, ec.GameID, currentHero, &ec.ActorID); err != nil {
		return err
	}
	return r.UpdatePlayerHero(ctx, ec.GameID, ec.ActorID, newHeroID)
}

func execShuffleHeroesDeck(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	return r.ReturnDiscardToHeroDeckAndShuffle(ctx, ec.GameID)
}

func execChangeMonster(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.MonsterSlot == nil {
		return errBadRequest
	}
	slot := *ec.MonsterSlot
	if slot != 0 && slot != 1 {
		return errBadRequest
	}
	newMonsterID, err := r.DrawRandomMonster(ctx)
	if err != nil {
		return err
	}
	return r.ReplaceActiveMonster(ctx, ec.GameID, slot, newMonsterID)
}

func execBlindDraw(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil {
		return err
	}
	if target.HeroID != nil {
		if err := r.AddHeroToDiscard(ctx, ec.GameID, *target.HeroID, ec.TargetPlayer); err != nil {
			return err
		}
	}
	newHeroID, err := r.DrawTopHeroFromDeck(ctx, ec.GameID)
	if err != nil {
		return err
	}
	return r.UpdatePlayerHero(ctx, ec.GameID, *ec.TargetPlayer, newHeroID)
}

func execForceTransform(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil || target.HeroID == nil {
		return errBadRequest
	}
	if err := r.AddHeroToDiscard(ctx, ec.GameID, *target.HeroID, ec.TargetPlayer); err != nil {
		return err
	}
	newHeroID, err := r.DrawTopHeroFromDeck(ctx, ec.GameID)
	if err != nil {
		return err
	}
	return r.UpdatePlayerHero(ctx, ec.GameID, *ec.TargetPlayer, newHeroID)
}

func execExecution(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	var payload struct {
		DiscardID *uuid.UUID `json:"discard_id"`
	}
	if len(ec.Payload) > 0 {
		_ = json.Unmarshal(ec.Payload, &payload)
	}
	if payload.DiscardID == nil {
		return errBadRequest
	}
	return r.BurnDiscardByID(ctx, ec.GameID, *payload.DiscardID)
}

func execCommandTheDead(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.MonsterSlot == nil {
		return errBadRequest
	}
	slot := *ec.MonsterSlot
	if slot != 0 && slot != 1 {
		return errBadRequest
	}
	lastDiscard, err := r.GetLastDiscardHeroID(ctx, ec.GameID)
	if err != nil || lastDiscard == nil || *lastDiscard == "" {
		return errBadRequest
	}
	hero, err := r.GetHeroByID(ctx, *lastDiscard)
	if err != nil {
		return err
	}
	monsterID, err := r.GetActiveMonsterAtSlot(ctx, ec.GameID, slot)
	if err != nil {
		return err
	}
	monster, err := r.GetMonsterByID(ctx, monsterID)
	if err != nil {
		return err
	}
	if hero.Strength > monster.Strength {
		if err := r.AddPlayerCoins(ctx, ec.GameID, ec.ActorID, monster.LootCoins); err != nil {
			return err
		}
		if err := r.AddPlayerGems(ctx, ec.GameID, ec.ActorID, monster.LootGems); err != nil {
			return err
		}
		r.EmitEvent(ctx, &Event{
			Kind:         EventKindGemsGained,
			GemsGameID:   ec.GameID,
			GemsPlayerID: ec.ActorID,
			GemsAmount:   monster.LootGems,
			GemsIsBonus:  false,
		})
		if monsterID == "ghost_king" {
			actor, _ := r.GetPlayerByID(ctx, ec.GameID, ec.ActorID)
			if actor != nil && actor.HeroID != nil && *actor.HeroID == "queen" {
				_ = r.AddPlayerCoins(ctx, ec.GameID, ec.ActorID, 1)
				_ = r.AddPlayerGems(ctx, ec.GameID, ec.ActorID, 3)
			}
		}
		_ = r.RemoveActiveMonsterSlot(ctx, ec.GameID, slot)
	} else {
		if err := r.AddPlayerLife(ctx, ec.GameID, ec.ActorID, -1); err != nil {
			return err
		}
		after, _ := r.GetPlayerByID(ctx, ec.GameID, ec.ActorID)
		if after != nil && after.Life <= 0 {
			_ = r.SetPlayerEliminated(ctx, ec.GameID, ec.ActorID, true)
		}
	}
	return nil
}
