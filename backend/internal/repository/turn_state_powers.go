package repository

import (
	"context"

	"github.com/google/uuid"
)

func (db *DB) SetChallengerBonus(ctx context.Context, gameID uuid.UUID, playerID *uuid.UUID, bonus int) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_turn_state SET challenger_player_id = $1, challenger_strength_bonus = $2 WHERE game_id = $3`,
		playerID, bonus, gameID,
	)
	return err
}

func (db *DB) GetChallengerBonus(ctx context.Context, gameID uuid.UUID) (playerID *uuid.UUID, bonus int, err error) {
	err = db.Pool.QueryRow(ctx,
		`SELECT challenger_player_id, COALESCE(challenger_strength_bonus, 0) FROM game_turn_state WHERE game_id = $1`,
		gameID,
	).Scan(&playerID, &bonus)
	return playerID, bonus, err
}

func (db *DB) SetDualAttackUsed(ctx context.Context, gameID uuid.UUID, used bool) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_turn_state SET dual_attack_used = $1 WHERE game_id = $2`,
		used, gameID,
	)
	return err
}

func (db *DB) GetDualAttackUsed(ctx context.Context, gameID uuid.UUID) (bool, error) {
	var used bool
	err := db.Pool.QueryRow(ctx,
		`SELECT COALESCE(dual_attack_used, false) FROM game_turn_state WHERE game_id = $1`,
		gameID,
	).Scan(&used)
	return used, err
}

func (db *DB) SetLastShootCorrect(ctx context.Context, gameID, playerID uuid.UUID, correct bool) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_turn_state SET last_shoot_correct_player_id = $1, last_shoot_correct = $2 WHERE game_id = $3`,
		playerID, correct, gameID,
	)
	return err
}

func (db *DB) GetLastShootCorrect(ctx context.Context, gameID, playerID uuid.UUID) (bool, error) {
	var storedID *uuid.UUID
	var correct *bool
	err := db.Pool.QueryRow(ctx,
		`SELECT last_shoot_correct_player_id, last_shoot_correct FROM game_turn_state WHERE game_id = $1`,
		gameID,
	).Scan(&storedID, &correct)
	if err != nil {
		return false, err
	}
	if storedID == nil || *storedID != playerID {
		return false, nil
	}
	if correct == nil {
		return false, nil
	}
	return *correct, nil
}

func (db *DB) SetMimicHero(ctx context.Context, gameID uuid.UUID, actorID *uuid.UUID, heroID *string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_turn_state SET mimicked_actor_id = $1, mimicked_hero_id = $2 WHERE game_id = $3`,
		actorID, heroID, gameID,
	)
	return err
}

func (db *DB) GetMimicHero(ctx context.Context, gameID uuid.UUID) (actorID *uuid.UUID, heroID *string, err error) {
	err = db.Pool.QueryRow(ctx,
		`SELECT mimicked_actor_id, mimicked_hero_id FROM game_turn_state WHERE game_id = $1`,
		gameID,
	).Scan(&actorID, &heroID)
	return actorID, heroID, err
}
