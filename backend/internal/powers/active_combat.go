package powers

import (
	"context"
	"encoding/json"
)

const werewolfHeroID = "werewolf"

func RegisterActiveCombat(reg *Registry) {
	reg.Register("fight_monster", PowerDef{Name: "fight_monster", CostGems: 1, TargetType: TargetMonster}, execFightMonster)
	reg.Register("fight_player_with_discarded_card", PowerDef{Name: "fight_player_with_discarded_card", CostGems: 0, TargetType: TargetOtherPlayer}, execFightPlayerWithDiscardedCard)
	reg.Register("add_strength_to_challenger", PowerDef{Name: "add_strength_to_challenger", CostGems: 2, TargetType: TargetSession}, execAddStrengthToChallenger)
	reg.Register("shoot_player", PowerDef{Name: "shoot_player", CostGems: 0, TargetType: TargetOtherPlayer}, execShootPlayer)
	reg.Register("all_in", PowerDef{Name: "all_in", CostGems: 69, TargetType: TargetSelf}, execAllIn)
	reg.Register("dual_attack", PowerDef{Name: "dual_attack", CostGems: 6, TargetType: TargetSession}, execDualAttack)
	reg.Register("see_player_card", PowerDef{Name: "see_player_card", CostGems: 0, TargetType: TargetOtherPlayer}, execSeePlayerCard)
}

func effectiveStrength(r PowerRunner, ctx context.Context, heroID string, isWolfForm bool) (int, error) {
	if heroID == werewolfHeroID {
		if isWolfForm {
			return 12, nil
		}
		return 5, nil
	}
	hero, err := r.GetHeroByID(ctx, heroID)
	if err != nil {
		return 0, err
	}
	return hero.Strength, nil
}

func execFightMonster(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.MonsterSlot == nil {
		return errBadRequest
	}
	slot := *ec.MonsterSlot
	if slot != 0 && slot != 1 {
		return errBadRequest
	}
	actor, err := r.GetPlayerByID(ctx, ec.GameID, ec.ActorID)
	if err != nil {
		return err
	}
	declared := actor.DeclaredHeroID
	if declared == nil {
		declared = actor.HeroID
	}
	if declared == nil {
		return errBadRequest
	}
	strength, err := effectiveStrength(r, ctx, *declared, actor.IsWolfForm)
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
	if strength > monster.Strength {
		if !HasRoyalImmunity(r, ctx, ec.GameID, ec.ActorID) {
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

func execFightPlayerWithDiscardedCard(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	var payload struct {
		DiscardedHeroID string `json:"discarded_hero_id"`
	}
	if len(ec.Payload) > 0 {
		_ = json.Unmarshal(ec.Payload, &payload)
	}
	if payload.DiscardedHeroID == "" {
		return errBadRequest
	}
	discardStrength, err := r.GetHeroByID(ctx, payload.DiscardedHeroID)
	if err != nil {
		return err
	}
	challengerID, bonus, _ := r.GetChallengerBonus(ctx, ec.GameID)
	actorStrength := discardStrength.Strength
	if challengerID != nil && *challengerID == ec.ActorID && bonus > 0 {
		actorStrength += bonus
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil || target.HeroID == nil {
		return errBadRequest
	}
	targetStrength, err := effectiveStrength(r, ctx, *target.HeroID, target.IsWolfForm)
	if err != nil {
		return err
	}
	if actorStrength > targetStrength {
		if err := r.AddPlayerLife(ctx, ec.GameID, *ec.TargetPlayer, -1); err != nil {
			return err
		}
		after, _ := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
		if after != nil && after.Life <= 0 {
			_ = r.SetPlayerEliminated(ctx, ec.GameID, *ec.TargetPlayer, true)
		}
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

func execAddStrengthToChallenger(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	return r.SetChallengerBonus(ctx, ec.GameID, &ec.ActorID, 1)
}

func execShootPlayer(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	if HasRoyalImmunity(r, ctx, ec.GameID, *ec.TargetPlayer) {
		return nil
	}
	var payload struct {
		HeroID string `json:"hero_id"`
	}
	if len(ec.Payload) > 0 {
		_ = json.Unmarshal(ec.Payload, &payload)
	}
	if payload.HeroID == "" {
		return errBadRequest
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil || target.HeroID == nil {
		return errBadRequest
	}
	correct := *target.HeroID == payload.HeroID
	if err := r.SetLastShootCorrect(ctx, ec.GameID, ec.ActorID, correct); err != nil {
		return err
	}
	if correct {
		return r.AddPlayerLife(ctx, ec.GameID, *ec.TargetPlayer, -1)
	}
	return nil
}

func execAllIn(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	correct, err := r.GetLastShootCorrect(ctx, ec.GameID, ec.ActorID)
	if err != nil {
		return err
	}
	actor, err := r.GetPlayerByID(ctx, ec.GameID, ec.ActorID)
	if err != nil {
		return err
	}
	if correct {
		return r.AddPlayerGems(ctx, ec.GameID, ec.ActorID, actor.Gems)
	}
	return r.AddPlayerGems(ctx, ec.GameID, ec.ActorID, -actor.Gems)
}

func execDualAttack(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	return r.SetDualAttackUsed(ctx, ec.GameID, true)
}

func execSeePlayerCard(ctx context.Context, ec ExecutionContext, r PowerRunner) error {
	if ec.TargetPlayer == nil {
		return errBadRequest
	}
	target, err := r.GetPlayerByID(ctx, ec.GameID, *ec.TargetPlayer)
	if err != nil || target.HeroID == nil {
		return errBadRequest
	}
	r.SecretReveal(ctx, ec.GameID, ec.ActorID, *target.HeroID)
	return nil
}
