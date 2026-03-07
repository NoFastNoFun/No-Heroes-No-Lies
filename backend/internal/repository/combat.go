package repository

import (
	"context"

	"github.com/google/uuid"
)

func (db *DB) AddPlayerLife(ctx context.Context, gameID, playerID uuid.UUID, delta int) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_players SET life = life + $1 WHERE game_id = $2 AND player_id = $3`,
		delta, gameID, playerID,
	)
	return err
}

func (db *DB) AddPlayerCoins(ctx context.Context, gameID, playerID uuid.UUID, delta int) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_players SET coins = coins + $1 WHERE game_id = $2 AND player_id = $3`,
		delta, gameID, playerID,
	)
	return err
}

func (db *DB) AddPlayerGems(ctx context.Context, gameID, playerID uuid.UUID, delta int) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_players SET gems = gems + $1 WHERE game_id = $2 AND player_id = $3`,
		delta, gameID, playerID,
	)
	return err
}

func (db *DB) SetPlayerEliminated(ctx context.Context, gameID, playerID uuid.UUID, eliminated bool) error {
	if !eliminated {
		_, err := db.Pool.Exec(ctx,
			`UPDATE game_players SET is_eliminated = false WHERE game_id = $1 AND player_id = $2`,
			gameID, playerID,
		)
		return err
	}
	var heroID *string
	err := db.Pool.QueryRow(ctx,
		`SELECT hero_id FROM game_players WHERE game_id = $1 AND player_id = $2`,
		gameID, playerID,
	).Scan(&heroID)
	if err != nil {
		return err
	}
	if heroID != nil && *heroID != "" {
		if err := db.AddHeroToDiscard(ctx, gameID, *heroID, nil); err != nil {
			return err
		}
	}
	_, err = db.Pool.Exec(ctx,
		`UPDATE game_players SET is_eliminated = true WHERE game_id = $1 AND player_id = $2`,
		gameID, playerID,
	)
	return err
}

func (db *DB) DrawTopMonsterFromDeck(ctx context.Context, gameID uuid.UUID) (monsterID string, err error) {
	err = db.Pool.QueryRow(ctx,
		`WITH top AS (
			SELECT monster_id, position FROM game_monster_deck WHERE game_id = $1 ORDER BY position LIMIT 1
		 )
		 DELETE FROM game_monster_deck WHERE game_id = $1 AND position = (SELECT position FROM top)
		 RETURNING monster_id`,
		gameID,
	).Scan(&monsterID)
	return monsterID, err
}

func (db *DB) GetActiveMonsterAtSlot(ctx context.Context, gameID uuid.UUID, slotIndex int) (monsterID string, err error) {
	err = db.Pool.QueryRow(ctx,
		`SELECT monster_id FROM game_active_monsters WHERE game_id = $1 AND slot_index = $2`,
		gameID, slotIndex,
	).Scan(&monsterID)
	return monsterID, err
}

func (db *DB) ReplaceActiveMonster(ctx context.Context, gameID uuid.UUID, slotIndex int, monsterID string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_active_monsters SET monster_id = $1 WHERE game_id = $2 AND slot_index = $3`,
		monsterID, gameID, slotIndex,
	)
	return err
}

func (db *DB) RemoveActiveMonsterSlot(ctx context.Context, gameID uuid.UUID, slotIndex int) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM game_active_monsters WHERE game_id = $1 AND slot_index = $2`,
		gameID, slotIndex,
	)
	return err
}

func (db *DB) InsertActiveMonster(ctx context.Context, gameID uuid.UUID, slotIndex int, monsterID string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO game_active_monsters (game_id, monster_id, slot_index) VALUES ($1, $2, $3)`,
		gameID, monsterID, slotIndex,
	)
	return err
}

func (db *DB) GetMonsterDeckSize(ctx context.Context, gameID uuid.UUID) (int, error) {
	var n int
	err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM game_monster_deck WHERE game_id = $1`, gameID).Scan(&n)
	return n, err
}

func (db *DB) PrependMonsterToDeck(ctx context.Context, gameID uuid.UUID, monsterID string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO game_monster_deck (game_id, monster_id, position)
		 SELECT $1, $2, COALESCE(MIN(position), 0) - 1 FROM game_monster_deck WHERE game_id = $1`,
		gameID, monsterID,
	)
	return err
}
