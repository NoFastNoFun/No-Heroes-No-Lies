package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/config"
	"github.com/no-heroes-no-lies/backend/internal/gameloop"
	"github.com/no-heroes-no-lies/backend/internal/models"
	"github.com/no-heroes-no-lies/backend/internal/powers"
	"github.com/no-heroes-no-lies/backend/internal/repository"
)

type TurnProcessor struct {
	db       *repository.DB
	cfg      *config.Config
	powers   *powers.Registry
	notifier GameNotifier
}

func NewTurnProcessor(db *repository.DB, cfg *config.Config, powerReg *powers.Registry, notifier GameNotifier) *TurnProcessor {
	return &TurnProcessor{db: db, cfg: cfg, powers: powerReg, notifier: notifier}
}

func (p *TurnProcessor) checkWinCondition(ctx context.Context, gameID uuid.UUID) error {
	game, err := p.db.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}
	if game.Status != models.GameStatusPlaying || game.WinConditionMet {
		return nil
	}
	players, err := p.db.GetGamePlayers(ctx, gameID)
	if err != nil {
		return err
	}
	var alive []models.GamePlayer
	for _, gp := range players {
		if !gp.IsEliminated {
			alive = append(alive, gp)
		}
	}
	if len(alive) == 1 {
		if err := p.db.SetGameFinished(ctx, gameID, alive[0].PlayerID); err != nil {
			return err
		}
		if p.notifier != nil {
			p.notifier.NotifyStateUpdated(gameID)
		}
		return nil
	}
	coinTarget := game.MaxPlayers + 1
	for _, gp := range players {
		if gp.Coins >= coinTarget {
			if err := p.db.SetGameFinished(ctx, gameID, gp.PlayerID); err != nil {
				return err
			}
			if p.notifier != nil {
				p.notifier.NotifyStateUpdated(gameID)
			}
			return nil
		}
	}
	return nil
}

func (p *TurnProcessor) Process(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command) gameloop.ProcessResult {
	turnState, err := p.db.GetGameTurnStateWithPicked(ctx, gameID)
	if err != nil {
		return gameloop.ProcessResult{Err: err}
	}
	game, err := p.db.GetGameByID(ctx, gameID)
	if err != nil {
		return gameloop.ProcessResult{Err: err}
	}
	if game.Status != models.GameStatusPlaying {
		return gameloop.ProcessResult{}
	}

	switch cmd.Type {
	case gameloop.CmdPick:
		return gameloop.ProcessResult{Err: p.processPick(ctx, gameID, cmd, turnState)}
	case gameloop.CmdDiscard:
		return gameloop.ProcessResult{Err: p.processDiscard(ctx, gameID, cmd, turnState)}
	case gameloop.CmdDeclare:
		return p.processDeclare(ctx, gameID, cmd, turnState)
	case gameloop.CmdChallengeWindowClosed:
		return gameloop.ProcessResult{Err: p.processChallengeWindowClosed(ctx, gameID)}
	case gameloop.CmdPower:
		return gameloop.ProcessResult{Err: p.processPower(ctx, gameID, cmd, turnState)}
	case gameloop.CmdSkipPower:
		return gameloop.ProcessResult{Err: p.processSkipPower(ctx, gameID, cmd, turnState)}
	case gameloop.CmdAttack:
		return gameloop.ProcessResult{Err: p.processAttack(ctx, gameID, cmd, turnState)}
	case gameloop.CmdDemask:
		return gameloop.ProcessResult{Err: p.processDemask(ctx, gameID, cmd, turnState)}
	case gameloop.CmdAccuse:
		return gameloop.ProcessResult{Err: p.processAccuse(ctx, gameID, cmd, turnState)}
	default:
		return gameloop.ProcessResult{}
	}
}

func (p *TurnProcessor) processPick(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) error {
	if turnState.Phase != models.PhasePick {
		return ErrBadRequest
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID != cmd.PlayerID {
		return ErrForbidden
	}
	heroID, err := p.db.DrawTopHeroFromDeck(ctx, gameID)
	if err != nil {
		return err
	}
	challengeEnd := (*time.Time)(nil)
	if err := p.db.UpdateTurnStatePhase(ctx, gameID, models.PhaseDiscard, &heroID, challengeEnd); err != nil {
		return err
	}
	if p.notifier != nil {
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

func (p *TurnProcessor) processDiscard(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) error {
	if turnState.Phase != models.PhaseDiscard {
		return ErrBadRequest
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID != cmd.PlayerID {
		return ErrForbidden
	}
	if turnState.PickedHeroID == nil {
		return ErrBadRequest
	}
	var payload struct {
		KeepPicked bool `json:"keep_picked"`
	}
	_ = json.Unmarshal(cmd.Payload, &payload)
	players, err := p.db.GetGamePlayers(ctx, gameID)
	if err != nil {
		return err
	}
	var currentHeroID string
	for _, gp := range players {
		if gp.PlayerID == cmd.PlayerID && gp.HeroID != nil {
			currentHeroID = *gp.HeroID
			break
		}
	}
	var toDiscard string
	var newHeroID string
	if payload.KeepPicked {
		toDiscard = currentHeroID
		newHeroID = *turnState.PickedHeroID
	} else {
		toDiscard = *turnState.PickedHeroID
		newHeroID = currentHeroID
	}
	if err := p.db.AddHeroToDiscard(ctx, gameID, toDiscard, &cmd.PlayerID); err != nil {
		return err
	}
	if err := p.db.UpdateGameLastDiscard(ctx, gameID, &toDiscard); err != nil {
		return err
	}
	if err := p.db.UpdatePlayerHero(ctx, gameID, cmd.PlayerID, newHeroID); err != nil {
		return err
	}
	if err := p.db.UpdateTurnStatePhase(ctx, gameID, models.PhaseDeclare, nil, nil); err != nil {
		return err
	}
	if p.notifier != nil {
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

func (p *TurnProcessor) processDeclare(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) gameloop.ProcessResult {
	if turnState.Phase != models.PhaseDeclare {
		return gameloop.ProcessResult{Err: ErrBadRequest}
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID != cmd.PlayerID {
		return gameloop.ProcessResult{Err: ErrForbidden}
	}
	var payload struct {
		HeroID              string `json:"hero_id"`
		IntendedAction      string `json:"intended_action"`
		IntendedMonsterSlot *int   `json:"intended_monster_slot"`
	}
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil || payload.HeroID == "" {
		return gameloop.ProcessResult{Err: ErrBadRequest}
	}
	if err := p.db.UpdatePlayerDeclaredHero(ctx, gameID, cmd.PlayerID, payload.HeroID); err != nil {
		return gameloop.ProcessResult{Err: err}
	}
	if payload.IntendedAction == "attack" && payload.IntendedMonsterSlot != nil {
		slot := *payload.IntendedMonsterSlot
		if slot == 0 || slot == 1 {
			monsterID, err := p.db.GetActiveMonsterAtSlot(ctx, gameID, slot)
			if err == nil && monsterID == "ghost_king" {
				nilTime := (*time.Time)(nil)
				if err := p.db.UpdateTurnStatePhase(ctx, gameID, models.PhaseAttack, nil, nilTime); err != nil {
					return gameloop.ProcessResult{Err: err}
				}
				if p.notifier != nil {
					p.notifier.NotifyStateUpdated(gameID)
				}
				return gameloop.ProcessResult{}
			}
		}
	}
	endsAt := time.Now().Add(p.cfg.ChallengeWindow)
	if err := p.db.SetChallengeWindowEndsAt(ctx, gameID, &endsAt); err != nil {
		return gameloop.ProcessResult{Err: err}
	}
	if p.notifier != nil {
		p.notifier.NotifyChallengeWindowOpen(gameID, endsAt)
		p.notifier.NotifyStateUpdated(gameID)
	}
	return gameloop.ProcessResult{
		Schedule: []gameloop.DelayedCommand{{
			After:   p.cfg.ChallengeWindow,
			Command: gameloop.Command{Type: gameloop.CmdChallengeWindowClosed},
		}},
	}
}

func (p *TurnProcessor) processChallengeWindowClosed(ctx context.Context, gameID uuid.UUID) error {
	turnState, err := p.db.GetGameTurnStateWithPicked(ctx, gameID)
	if err != nil {
		return err
	}
	if turnState.ChallengeWindowEndsAt == nil {
		return nil
	}
	if err := p.db.UpdateTurnStatePhase(ctx, gameID, models.PhasePower, nil, nil); err != nil {
		return err
	}
	if p.notifier != nil {
		p.notifier.NotifyChallengeWindowClosed(gameID)
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

func (p *TurnProcessor) processPower(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) error {
	if turnState.Phase != models.PhasePower {
		return ErrBadRequest
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID != cmd.PlayerID {
		return ErrForbidden
	}
	var payload struct {
		PowerName         string     `json:"power_name"`
		TargetPlayerID    *uuid.UUID `json:"target_player_id"`
		TargetMonsterSlot *int       `json:"target_monster_slot"`
		DiscardID         *uuid.UUID `json:"discard_id"`
	}
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil || payload.PowerName == "" {
		return ErrBadRequest
	}
	def, ok := p.powers.Get(payload.PowerName)
	if !ok {
		return ErrBadRequest
	}
	if def.IsPassive {
		return ErrBadRequest
	}
	runner := NewPowerRunner(p.db, p.notifier, p.powers)
	actor, err := p.db.GetPlayerByID(ctx, gameID, cmd.PlayerID)
	if err != nil {
		return err
	}
	if actor.HeroID == nil {
		return ErrBadRequest
	}
	hero, err := p.db.GetHeroByID(ctx, *actor.HeroID)
	if err != nil {
		return err
	}
	if payload.PowerName != hero.Power1 && payload.PowerName != hero.Power2 {
		return ErrBadRequest
	}
	if actor.Gems < def.CostGems {
		return ErrBadRequest
	}
	if def.CostGems > 0 {
		if err := p.db.AddPlayerGems(ctx, gameID, cmd.PlayerID, -def.CostGems); err != nil {
			return err
		}
	}
	ec := powers.ExecutionContext{
		GameID:       gameID,
		ActorID:      cmd.PlayerID,
		TargetPlayer: payload.TargetPlayerID,
		MonsterSlot:  payload.TargetMonsterSlot,
		Payload:      cmd.Payload,
	}
	if payload.PowerName == hero.Power2 {
		if err := p.powers.Execute(ctx, hero.Power1, ec, runner); err != nil {
			if def.CostGems > 0 {
				_ = p.db.AddPlayerGems(ctx, gameID, cmd.PlayerID, def.CostGems)
			}
			return err
		}
	}
	if err := p.powers.Execute(ctx, payload.PowerName, ec, runner); err != nil {
		if def.CostGems > 0 {
			_ = p.db.AddPlayerGems(ctx, gameID, cmd.PlayerID, def.CostGems)
		}
		return err
	}
	_ = p.checkWinCondition(ctx, gameID)
	if err := p.db.UpdateTurnStatePhase(ctx, gameID, models.PhaseAttack, nil, nil); err != nil {
		return err
	}
	if p.notifier != nil {
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

func (p *TurnProcessor) processSkipPower(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) error {
	if turnState.Phase != models.PhasePower {
		return ErrBadRequest
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID != cmd.PlayerID {
		return ErrForbidden
	}
	if err := p.db.UpdateTurnStatePhase(ctx, gameID, models.PhaseAttack, nil, nil); err != nil {
		return err
	}
	if p.notifier != nil {
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

const werewolfHeroID = "werewolf"

func (p *TurnProcessor) effectiveDeclaredStrength(ctx context.Context, declaredHeroID string, isWolfForm bool) (int, error) {
	if declaredHeroID == werewolfHeroID {
		if isWolfForm {
			return 12, nil
		}
		return 5, nil
	}
	hero, err := p.db.GetHeroByID(ctx, declaredHeroID)
	if err != nil {
		return 0, err
	}
	return hero.Strength, nil
}

func (p *TurnProcessor) processAttack(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) error {
	if turnState.Phase != models.PhaseAttack {
		return ErrBadRequest
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID != cmd.PlayerID {
		return ErrForbidden
	}
	var payload struct {
		MonsterSlot *int `json:"monster_slot"`
	}
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil || payload.MonsterSlot == nil {
		return ErrBadRequest
	}
	slot := *payload.MonsterSlot
	if slot != 0 && slot != 1 {
		return ErrBadRequest
	}
	players, err := p.db.GetGamePlayers(ctx, gameID)
	if err != nil {
		return err
	}
	var actor *models.GamePlayer
	for i := range players {
		if players[i].PlayerID == cmd.PlayerID {
			actor = &players[i]
			break
		}
	}
	if actor == nil || actor.DeclaredHeroID == nil || *actor.DeclaredHeroID == "" {
		return ErrBadRequest
	}
	heroIDForStrength := *actor.DeclaredHeroID
	if mimicActorID, mimicHeroID, err := p.db.GetMimicHero(ctx, gameID); err == nil && mimicActorID != nil && *mimicActorID == cmd.PlayerID && mimicHeroID != nil {
		heroIDForStrength = *mimicHeroID
	}
	strength, err := p.effectiveDeclaredStrength(ctx, heroIDForStrength, actor.IsWolfForm)
	if err != nil {
		return err
	}
	monsterID, err := p.db.GetActiveMonsterAtSlot(ctx, gameID, slot)
	if err != nil {
		return ErrBadRequest
	}
	monster, err := p.db.GetMonsterByID(ctx, monsterID)
	if err != nil {
		return err
	}
	if strength > monster.Strength {
		royalImmunity := actor.HeroID != nil && *actor.HeroID == "queen"
		if !royalImmunity {
			if err := p.db.AddPlayerCoins(ctx, gameID, cmd.PlayerID, monster.LootCoins); err != nil {
				return err
			}
			if err := p.db.AddPlayerGems(ctx, gameID, cmd.PlayerID, monster.LootGems); err != nil {
				return err
			}
			p.emitGemsGainedForPassives(ctx, gameID, cmd.PlayerID, monster.LootGems)
		}
		if monsterID == "ghost_king" && actor.HeroID != nil {
			if *actor.HeroID == "prince" || *actor.HeroID == "queen" {
				_ = p.db.AddPlayerCoins(ctx, gameID, cmd.PlayerID, 2)
				_ = p.db.AddPlayerGems(ctx, gameID, cmd.PlayerID, 5)
				p.emitGemsGainedForPassives(ctx, gameID, cmd.PlayerID, 5)
			}
		}
		_ = p.checkWinCondition(ctx, gameID)
		_ = p.db.RemoveActiveMonsterSlot(ctx, gameID, slot)
	} else {
		if err := p.db.AddPlayerLife(ctx, gameID, cmd.PlayerID, -1); err != nil {
			return err
		}
		after, _ := p.db.GetPlayerByID(ctx, gameID, cmd.PlayerID)
		if after != nil && after.Life <= 0 {
			p.eliminatePlayer(ctx, gameID, cmd.PlayerID)
		}
		_ = p.checkWinCondition(ctx, gameID)
	}
	dualUsed, _ := p.db.GetDualAttackUsed(ctx, gameID)
	if dualUsed {
		_ = p.db.SetDualAttackUsed(ctx, gameID, false)
		if p.notifier != nil {
			p.notifier.NotifyStateUpdated(gameID)
		}
		return nil
	}
	nextID, err := p.db.AdvanceToNextTurn(ctx, gameID)
	if err != nil {
		return err
	}
	if nextID != uuid.Nil {
		runner := NewPowerRunner(p.db, p.notifier, p.powers)
		runner.EmitEvent(ctx, &powers.Event{
			Kind:              powers.EventKindTurnStart,
			TurnStartGameID:   gameID,
			TurnStartPlayerID: nextID,
		})
	}
	if p.notifier != nil {
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

func (p *TurnProcessor) processDemask(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) error {
	if turnState.Phase != models.PhaseAttack {
		return ErrBadRequest
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID != cmd.PlayerID {
		return ErrForbidden
	}
	var payload struct {
		TargetPlayerID *uuid.UUID `json:"target_player_id"`
		HeroID         string     `json:"hero_id"`
	}
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil || payload.TargetPlayerID == nil || payload.HeroID == "" {
		return ErrBadRequest
	}
	if *payload.TargetPlayerID == cmd.PlayerID {
		return ErrBadRequest
	}
	actor, err := p.db.GetPlayerByID(ctx, gameID, cmd.PlayerID)
	if err != nil {
		return err
	}
	if actor.Gems < 6 {
		return ErrBadRequest
	}
	target, err := p.db.GetPlayerByID(ctx, gameID, *payload.TargetPlayerID)
	if err != nil {
		return err
	}
	if target.HeroID == nil {
		return ErrBadRequest
	}
	if *target.HeroID != payload.HeroID {
		_ = p.checkWinCondition(ctx, gameID)
		nextID, err := p.db.AdvanceToNextTurn(ctx, gameID)
		if err != nil {
			return err
		}
		if nextID != uuid.Nil {
			runner := NewPowerRunner(p.db, p.notifier, p.powers)
			runner.EmitEvent(ctx, &powers.Event{
				Kind:              powers.EventKindTurnStart,
				TurnStartGameID:   gameID,
				TurnStartPlayerID: nextID,
			})
		}
		if p.notifier != nil {
			p.notifier.NotifyStateUpdated(gameID)
		}
		return nil
	}
	if err := p.db.AddHeroToDiscard(ctx, gameID, *target.HeroID, &cmd.PlayerID); err != nil {
		return err
	}
	newHeroID, drawErr := p.db.DrawTopHeroFromDeck(ctx, gameID)
	if drawErr == nil {
		_ = p.db.UpdatePlayerHero(ctx, gameID, *payload.TargetPlayerID, newHeroID)
	} else {
		_ = p.db.ClearPlayerHero(ctx, gameID, *payload.TargetPlayerID)
	}
	if err := p.db.AddPlayerLife(ctx, gameID, *payload.TargetPlayerID, -1); err != nil {
		return err
	}
	after, _ := p.db.GetPlayerByID(ctx, gameID, *payload.TargetPlayerID)
	if after != nil && after.Life <= 0 {
		p.eliminatePlayer(ctx, gameID, *payload.TargetPlayerID)
	}
	_ = p.checkWinCondition(ctx, gameID)
	nextID, err := p.db.AdvanceToNextTurn(ctx, gameID)
	if err != nil {
		return err
	}
	if nextID != uuid.Nil {
		runner := NewPowerRunner(p.db, p.notifier, p.powers)
		runner.EmitEvent(ctx, &powers.Event{
			Kind:              powers.EventKindTurnStart,
			TurnStartGameID:   gameID,
			TurnStartPlayerID: nextID,
		})
	}
	if p.notifier != nil {
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

func (p *TurnProcessor) processAccuse(ctx context.Context, gameID uuid.UUID, cmd gameloop.Command, turnState *models.GameTurnState) error {
	if turnState.Phase != models.PhaseDeclare {
		return ErrBadRequest
	}
	if turnState.ChallengeWindowEndsAt == nil {
		return ErrBadRequest
	}
	if turnState.ActingPlayerID == nil || *turnState.ActingPlayerID == cmd.PlayerID {
		return ErrForbidden
	}
	var payload struct {
		HeroID string `json:"hero_id"`
	}
	if err := json.Unmarshal(cmd.Payload, &payload); err != nil || payload.HeroID == "" {
		return ErrBadRequest
	}
	actingID := *turnState.ActingPlayerID
	acting, err := p.db.GetPlayerByID(ctx, gameID, actingID)
	if err != nil {
		return err
	}
	if acting.HeroID == nil {
		return ErrBadRequest
	}
	if *acting.HeroID != payload.HeroID {
		if err := p.db.AddPlayerLife(ctx, gameID, cmd.PlayerID, -1); err != nil {
			return err
		}
		accuserAfter, _ := p.db.GetPlayerByID(ctx, gameID, cmd.PlayerID)
		if accuserAfter != nil && accuserAfter.Life <= 0 {
			p.eliminatePlayer(ctx, gameID, cmd.PlayerID)
		}
		_ = p.checkWinCondition(ctx, gameID)
		nilTime := (*time.Time)(nil)
		if err := p.db.SetChallengeWindowEndsAt(ctx, gameID, nilTime); err != nil {
			return err
		}
		if err := p.db.UpdateTurnStatePhase(ctx, gameID, models.PhasePower, nil, nilTime); err != nil {
			return err
		}
		if p.notifier != nil {
			p.notifier.NotifyChallengeWindowClosed(gameID)
			p.notifier.NotifyStateUpdated(gameID)
		}
		return nil
	}
	if err := p.db.AddHeroToDiscard(ctx, gameID, *acting.HeroID, &cmd.PlayerID); err != nil {
		return err
	}
	newHeroID, drawErr := p.db.DrawTopHeroFromDeck(ctx, gameID)
	if drawErr == nil {
		_ = p.db.UpdatePlayerHero(ctx, gameID, actingID, newHeroID)
	} else {
		_ = p.db.ClearPlayerHero(ctx, gameID, actingID)
	}
	if err := p.db.UpdatePlayerDeclaredHero(ctx, gameID, actingID, ""); err != nil {
		return err
	}
	if err := p.db.AddPlayerLife(ctx, gameID, actingID, -1); err != nil {
		return err
	}
	after, _ := p.db.GetPlayerByID(ctx, gameID, actingID)
	if after != nil && after.Life <= 0 {
		p.eliminatePlayer(ctx, gameID, actingID)
	}
	_ = p.checkWinCondition(ctx, gameID)
	nilTime := (*time.Time)(nil)
	if err := p.db.SetChallengeWindowEndsAt(ctx, gameID, nilTime); err != nil {
		return err
	}
	nextID, err := p.db.AdvanceToNextTurn(ctx, gameID)
	if err != nil {
		return err
	}
	if nextID != uuid.Nil {
		runner := NewPowerRunner(p.db, p.notifier, p.powers)
		runner.EmitEvent(ctx, &powers.Event{
			Kind:              powers.EventKindTurnStart,
			TurnStartGameID:   gameID,
			TurnStartPlayerID: nextID,
		})
	}
	if p.notifier != nil {
		p.notifier.NotifyChallengeWindowClosed(gameID)
		p.notifier.NotifyStateUpdated(gameID)
	}
	return nil
}

func (p *TurnProcessor) emitGemsGainedForPassives(ctx context.Context, gameID, playerID uuid.UUID, amount int) {
	if amount <= 0 || p.powers == nil {
		return
	}
	runner := NewPowerRunner(p.db, p.notifier, p.powers)
	runner.EmitEvent(ctx, &powers.Event{
		Kind:         powers.EventKindGemsGained,
		GemsGameID:   gameID,
		GemsPlayerID: playerID,
		GemsAmount:   amount,
		GemsIsBonus:  false,
	})
}

func (p *TurnProcessor) eliminatePlayer(ctx context.Context, gameID, playerID uuid.UUID) {
	_ = p.db.SetPlayerEliminated(ctx, gameID, playerID, true)
}