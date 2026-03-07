package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/models"
)

func (db *DB) GetGameTurnState(ctx context.Context, gameID uuid.UUID) (*models.GameTurnState, error) {
	var s models.GameTurnState
	var challengeEnd *time.Time
	var pendingType, powerStep *string
	var payload []byte
	var actingID *uuid.UUID
	var pickedID *string
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

func (db *DB) GetActiveMonsters(ctx context.Context, gameID uuid.UUID) ([]models.ActiveMonsterView, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT slot_index, monster_id FROM game_active_monsters WHERE game_id = $1 ORDER BY slot_index`,
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ActiveMonsterView
	for rows.Next() {
		var m models.ActiveMonsterView
		if err := rows.Scan(&m.SlotIndex, &m.MonsterID); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (db *DB) GetHeroDeckSize(ctx context.Context, gameID uuid.UUID) (int, error) {
	var n int
	err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM game_hero_deck WHERE game_id = $1 AND burned = false`, gameID).Scan(&n)
	return n, err
}
