package repository

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/models"
)

func (db *DB) DrawTopHeroFromDeck(ctx context.Context, gameID uuid.UUID) (heroID string, err error) {
	err = db.Pool.QueryRow(ctx,
		`WITH top AS (
			SELECT hero_id, position FROM game_hero_deck WHERE game_id = $1 AND burned = false ORDER BY position LIMIT 1
		 )
		 DELETE FROM game_hero_deck WHERE game_id = $1 AND position = (SELECT position FROM top)
		 RETURNING hero_id`,
		gameID,
	).Scan(&heroID)
	return heroID, err
}

func (db *DB) AddHeroToDiscard(ctx context.Context, gameID uuid.UUID, heroID string, discardedBy *uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO game_hero_discard (game_id, hero_id, discarded_by_player_id) VALUES ($1, $2, $3)`,
		gameID, heroID, discardedBy,
	)
	return err
}

func (db *DB) UpdateGameLastDiscard(ctx context.Context, gameID uuid.UUID, heroID *string) error {
	_, err := db.Pool.Exec(ctx, `UPDATE games SET last_discard_hero_id = $1, updated_at = now() WHERE id = $2`, heroID, gameID)
	return err
}

func (db *DB) UpdateTurnStatePhase(ctx context.Context, gameID uuid.UUID, phase models.TurnPhase, pickedHeroID *string, challengeEnd *time.Time) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_turn_state SET phase = $1, phase_entered_at = now(), challenge_window_ends_at = $2, picked_hero_id = $3,
		 pending_action_type = NULL, pending_action_payload = NULL WHERE game_id = $4`,
		phase, challengeEnd, pickedHeroID, gameID,
	)
	return err
}

func (db *DB) UpdatePlayerHero(ctx context.Context, gameID, playerID uuid.UUID, heroID string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_players SET hero_id = $1 WHERE game_id = $2 AND player_id = $3`,
		heroID, gameID, playerID,
	)
	return err
}

func (db *DB) ClearPlayerHero(ctx context.Context, gameID, playerID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_players SET hero_id = NULL WHERE game_id = $1 AND player_id = $2`,
		gameID, playerID,
	)
	return err
}

func (db *DB) SetChallengeWindowEndsAt(ctx context.Context, gameID uuid.UUID, endsAt *time.Time) error {
	_, err := db.Pool.Exec(ctx, `UPDATE game_turn_state SET challenge_window_ends_at = $1 WHERE game_id = $2`, endsAt, gameID)
	return err
}

func (db *DB) UpdatePlayerDeclaredHero(ctx context.Context, gameID, playerID uuid.UUID, heroID string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_players SET declared_hero_id = $1 WHERE game_id = $2 AND player_id = $3`,
		heroID, gameID, playerID,
	)
	return err
}

type DiscardRow struct {
	ID     uuid.UUID
	HeroID string
}

func (db *DB) GetDiscardPile(ctx context.Context, gameID uuid.UUID) ([]DiscardRow, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, hero_id FROM game_hero_discard WHERE game_id = $1 AND (burned = false OR burned IS NULL)`,
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DiscardRow
	for rows.Next() {
		var r DiscardRow
		if err := rows.Scan(&r.ID, &r.HeroID); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (db *DB) BurnDiscardByID(ctx context.Context, gameID uuid.UUID, discardID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_hero_discard SET burned = true WHERE game_id = $1 AND id = $2`,
		gameID, discardID,
	)
	return err
}

func (db *DB) ReturnDiscardToHeroDeckAndShuffle(ctx context.Context, gameID uuid.UUID) error {
	rows, err := db.Pool.Query(ctx,
		`SELECT hero_id FROM game_hero_discard WHERE game_id = $1 AND (burned = false OR burned IS NULL)`,
		gameID,
	)
	if err != nil {
		return err
	}
	var heroIDs []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			rows.Close()
			return err
		}
		heroIDs = append(heroIDs, h)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(heroIDs) == 0 {
		return nil
	}
	_, err = db.Pool.Exec(ctx, `DELETE FROM game_hero_discard WHERE game_id = $1 AND (burned = false OR burned IS NULL)`, gameID)
	if err != nil {
		return err
	}
	rand.Shuffle(len(heroIDs), func(i, j int) { heroIDs[i], heroIDs[j] = heroIDs[j], heroIDs[i] })
	var maxPos int
	err = db.Pool.QueryRow(ctx, `SELECT COALESCE(MAX(position), -1) FROM game_hero_deck WHERE game_id = $1`, gameID).Scan(&maxPos)
	if err != nil {
		return err
	}
	for i, heroID := range heroIDs {
		_, err = db.Pool.Exec(ctx,
			`INSERT INTO game_hero_deck (game_id, hero_id, position, burned) VALUES ($1, $2, $3, false)`,
			gameID, heroID, maxPos+1+i,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) GetGameTurnStateWithPicked(ctx context.Context, gameID uuid.UUID) (*models.GameTurnState, error) {
	var s models.GameTurnState
	var challengeEnd *time.Time
	var pendingType, powerStep, pickedID *string
	var payload []byte
	var actingID *uuid.UUID
	err := db.Pool.QueryRow(ctx,
		`SELECT game_id, phase, phase_entered_at, challenge_window_ends_at, pending_action_type, pending_action_payload, acting_player_id, power_step, picked_hero_id
		 FROM game_turn_state WHERE game_id = $1`,
		gameID,
	).Scan(&s.GameID, &s.Phase, &s.PhaseEnteredAt, &challengeEnd, &pendingType, &payload, &actingID, &powerStep, &pickedID)
	if err != nil {
		return nil, err
	}
	s.ChallengeWindowEndsAt = challengeEnd
	s.PendingActionType = pendingType
	s.PendingActionPayload = payload
	s.ActingPlayerID = actingID
	s.PowerStep = powerStep
	s.PickedHeroID = pickedID
	return &s, nil
}

func (db *DB) AdvanceToNextTurn(ctx context.Context, gameID uuid.UUID) (nextPlayerID uuid.UUID, err error) {
	players, err := db.GetGamePlayers(ctx, gameID)
	if err != nil {
		return uuid.Nil, err
	}
	if len(players) == 0 {
		return uuid.Nil, nil
	}
	var turnIndex int
	err = db.Pool.QueryRow(ctx, `SELECT current_turn_index FROM games WHERE id = $1`, gameID).Scan(&turnIndex)
	if err != nil {
		return uuid.Nil, err
	}
	nextIndex := (turnIndex + 1) % len(players)
	for i := 0; i < len(players); i++ {
		if !players[nextIndex].IsEliminated {
			nextPlayerID = players[nextIndex].PlayerID
			break
		}
		nextIndex = (nextIndex + 1) % len(players)
	}
	if nextPlayerID == uuid.Nil {
		return uuid.Nil, nil
	}
	_, err = db.Pool.Exec(ctx, `UPDATE games SET current_turn_index = $1, updated_at = now() WHERE id = $2`, nextIndex, gameID)
	if err != nil {
		return uuid.Nil, err
	}
	_, err = db.Pool.Exec(ctx,
		`UPDATE game_turn_state SET phase = 'pick', phase_entered_at = now(), acting_player_id = $1, challenge_window_ends_at = NULL, picked_hero_id = NULL,
		 challenger_player_id = NULL, challenger_strength_bonus = 0, dual_attack_used = false, last_shoot_correct_player_id = NULL, last_shoot_correct = NULL, mimicked_actor_id = NULL, mimicked_hero_id = NULL WHERE game_id = $2`,
		nextPlayerID, gameID,
	)
	if err != nil {
		return uuid.Nil, err
	}
	if err := db.fillEmptyActiveMonsterSlots(ctx, gameID); err != nil {
		return uuid.Nil, err
	}
	return nextPlayerID, nil
}

func (db *DB) fillEmptyActiveMonsterSlots(ctx context.Context, gameID uuid.UUID) error {
	for slot := 0; slot <= 1; slot++ {
		_, err := db.GetActiveMonsterAtSlot(ctx, gameID, slot)
		if err == nil {
			continue
		}
		monsterID, drawErr := db.DrawRandomMonsterByWeight(ctx)
		if drawErr != nil {
			continue
		}
		_ = db.InsertActiveMonster(ctx, gameID, slot, monsterID)
	}
	return nil
}
