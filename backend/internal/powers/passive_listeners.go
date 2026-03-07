package powers

import (
	"context"

	"github.com/google/uuid"
)

const (
	heroIDKeepGems        = "geant"
	heroIDTeamwork        = "chevalier"
	heroIDTribute         = "prince"
	heroIDAlternateStrength = "werewolf"
	heroIDRoyalImmunity   = "queen"
)

func RegisterPassiveListeners() {
	RegisterEventHandler(EventKindStealAttempt, handleStealAttempt)
	RegisterEventHandler(EventKindGemsGained, handleGemsGained)
	RegisterEventHandler(EventKindTurnStart, handleTurnStart)
}

func handleStealAttempt(ctx context.Context, ev *Event) {
	if ev.Cancelled {
		return
	}
	if ev.Runner == nil {
		return
	}
	target, err := ev.Runner.GetPlayerByID(ctx, ev.StealGameID, ev.TargetID)
	if err != nil {
		return
	}
	declaredID := target.DeclaredHeroID
	if declaredID == nil || *declaredID == "" {
		return
	}
	hero, err := ev.Runner.GetHeroByID(ctx, *declaredID)
	if err != nil {
		return
	}
	if hero.Power1 == "keep_gems" || hero.Power2 == "keep_gems" {
		ev.Cancelled = true
	}
}

func handleGemsGained(ctx context.Context, ev *Event) {
	if ev.GemsIsBonus || ev.Runner == nil {
		return
	}
	gameID := ev.GemsGameID
	playerID := ev.GemsPlayerID
	amount := ev.GemsAmount

	player, err := ev.Runner.GetPlayerByID(ctx, gameID, playerID)
	if err != nil {
		return
	}
	totalGain := amount
	declaredID := player.DeclaredHeroID
	if declaredID != nil && *declaredID != "" {
		hero, err := ev.Runner.GetHeroByID(ctx, *declaredID)
		if err == nil && (hero.Power1 == "teamwork" || hero.Power2 == "teamwork") {
			_ = ev.Runner.AddPlayerGems(ctx, gameID, playerID, amount)
			totalGain = amount * 2
		}
	}

	players, err := ev.Runner.GetGamePlayers(ctx, gameID)
	if err != nil {
		return
	}
	for _, p := range players {
		if p.PlayerID == playerID {
			continue
		}
		declared := p.DeclaredHeroID
		if declared == nil || *declared == "" {
			continue
		}
		h, err := ev.Runner.GetHeroByID(ctx, *declared)
		if err != nil {
			continue
		}
		if h.Power1 == "tribute" || h.Power2 == "tribute" {
			half := totalGain / 2
			if half > 0 {
				_ = ev.Runner.AddPlayerGems(ctx, gameID, p.PlayerID, half)
				_ = ev.Runner.AddPlayerGems(ctx, gameID, playerID, -half)
			}
			break
		}
	}
}

func handleTurnStart(ctx context.Context, ev *Event) {
	if ev.Runner == nil {
		return
	}
	gameID := ev.TurnStartGameID
	playerID := ev.TurnStartPlayerID
	player, err := ev.Runner.GetPlayerByID(ctx, gameID, playerID)
	if err != nil || player.HeroID == nil {
		return
	}
	if *player.HeroID != heroIDAlternateStrength {
		return
	}
	_ = ev.Runner.SetPlayerWolfForm(ctx, gameID, playerID, !player.IsWolfForm)
}

func HasRoyalImmunity(runner PowerRunner, ctx context.Context, gameID, playerID uuid.UUID) bool {
	if runner == nil {
		return false
	}
	p, err := runner.GetPlayerByID(ctx, gameID, playerID)
	if err != nil || p.HeroID == nil {
		return false
	}
	return *p.HeroID == heroIDRoyalImmunity
}
