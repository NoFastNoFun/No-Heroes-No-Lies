package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/models"
)

func (db *DB) RunGameSetup(ctx context.Context, gameID uuid.UUID, firstPlayerID uuid.UUID, heroDeck []string, activeMonsters [2]string, playerHeroes map[uuid.UUID]string) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, heroID := range heroDeck {
		_, err := tx.Exec(ctx,
			`INSERT INTO game_hero_deck (game_id, hero_id, position, burned) VALUES ($1, $2, $3, false)`,
			gameID, heroID, i,
		)
		if err != nil {
			return err
		}
	}
	for slot, monsterID := range activeMonsters {
		_, err := tx.Exec(ctx,
			`INSERT INTO game_active_monsters (game_id, monster_id, slot_index) VALUES ($1, $2, $3)`,
			gameID, monsterID, slot,
		)
		if err != nil {
			return err
		}
	}
	for playerID, heroID := range playerHeroes {
		_, err := tx.Exec(ctx,
			`UPDATE game_players SET hero_id = $1 WHERE game_id = $2 AND player_id = $3`,
			heroID, gameID, playerID,
		)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO game_turn_state (game_id, phase, acting_player_id) VALUES ($1, 'pick', $2)`,
		gameID, firstPlayerID,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE games SET status = $1, updated_at = now() WHERE id = $2`, models.GameStatusPlaying, gameID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
