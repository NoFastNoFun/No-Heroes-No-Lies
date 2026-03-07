package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/models"
)

func (db *DB) CreateGame(ctx context.Context, creatorID uuid.UUID, maxPlayers int) (*models.Game, error) {
	g := &models.Game{
		CreatorID: creatorID,
		Status:    models.GameStatusLobby,
		MaxPlayers: maxPlayers,
	}
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO games (creator_id, status, max_players)
		 VALUES ($1, 'lobby', $2)
		 RETURNING id, creator_id, status, max_players, current_turn_index, round_number, win_condition_met, created_at, updated_at`,
		creatorID, maxPlayers,
	).Scan(&g.ID, &g.CreatorID, &g.Status, &g.MaxPlayers, &g.CurrentTurnIndex, &g.RoundNumber, &g.WinConditionMet, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_, err = db.Pool.Exec(ctx,
		`INSERT INTO game_players (game_id, player_id, slot_index, is_game_master) VALUES ($1, $2, 0, true)`,
		g.ID, creatorID,
	)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (db *DB) GetGameByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	var g models.Game
	var lastDiscard *string
	var winnerID *uuid.UUID
	err := db.Pool.QueryRow(ctx,
		`SELECT id, creator_id, status, max_players, current_turn_index, round_number, win_condition_met, winner_id, last_discard_hero_id, created_at, updated_at
		 FROM games WHERE id = $1`,
		id,
	).Scan(&g.ID, &g.CreatorID, &g.Status, &g.MaxPlayers, &g.CurrentTurnIndex, &g.RoundNumber, &g.WinConditionMet, &winnerID, &lastDiscard, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	g.WinnerID = winnerID
	g.LastDiscardHeroID = lastDiscard
	return &g, nil
}

func (db *DB) ListGamesByStatus(ctx context.Context, status models.GameStatus) ([]models.GameListItem, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT g.id, g.creator_id, g.status, g.max_players, g.created_at,
		        COUNT(gp.player_id)::int AS player_count
		 FROM games g
		 LEFT JOIN game_players gp ON gp.game_id = g.id
		 WHERE g.status = $1
		 GROUP BY g.id, g.creator_id, g.status, g.max_players, g.created_at
		 ORDER BY g.created_at DESC`,
		status,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.GameListItem
	for rows.Next() {
		var item models.GameListItem
		err := rows.Scan(&item.ID, &item.CreatorID, &item.Status, &item.MaxPlayers, &item.CreatedAt, &item.PlayerCount)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (db *DB) GetGamePlayers(ctx context.Context, gameID uuid.UUID) ([]models.GamePlayer, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT game_id, player_id, slot_index, is_game_master, life, coins, gems, hero_id, declared_hero_id, is_wolf_form, is_eliminated, joined_at
		 FROM game_players WHERE game_id = $1 ORDER BY slot_index`,
		gameID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.GamePlayer
	for rows.Next() {
		var gp models.GamePlayer
		err := rows.Scan(&gp.GameID, &gp.PlayerID, &gp.SlotIndex, &gp.IsGameMaster, &gp.Life, &gp.Coins, &gp.Gems, &gp.HeroID, &gp.DeclaredHeroID, &gp.IsWolfForm, &gp.IsEliminated, &gp.JoinedAt)
		if err != nil {
			return nil, err
		}
		out = append(out, gp)
	}
	return out, rows.Err()
}

func (db *DB) JoinGame(ctx context.Context, gameID, playerID uuid.UUID) (slotIndex int, err error) {
	var nextSlot int
	err = db.Pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(slot_index), -1) + 1 FROM game_players WHERE game_id = $1`,
		gameID,
	).Scan(&nextSlot)
	if err != nil {
		return 0, err
	}
	_, err = db.Pool.Exec(ctx,
		`INSERT INTO game_players (game_id, player_id, slot_index) VALUES ($1, $2, $3)`,
		gameID, playerID, nextSlot,
	)
	return nextSlot, err
}

func (db *DB) GamePlayerCount(ctx context.Context, gameID uuid.UUID) (int, error) {
	var n int
	err := db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM game_players WHERE game_id = $1`, gameID).Scan(&n)
	return n, err
}

func (db *DB) IsPlayerInGame(ctx context.Context, gameID, playerID uuid.UUID) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM game_players WHERE game_id = $1 AND player_id = $2)`,
		gameID, playerID,
	).Scan(&exists)
	return exists, err
}

func (db *DB) IsGameMaster(ctx context.Context, gameID, playerID uuid.UUID) (bool, error) {
	var ok bool
	err := db.Pool.QueryRow(ctx,
		`SELECT is_game_master FROM game_players WHERE game_id = $1 AND player_id = $2`,
		gameID, playerID,
	).Scan(&ok)
	return ok, err
}

func (db *DB) SetGameStatus(ctx context.Context, gameID uuid.UUID, status models.GameStatus) error {
	_, err := db.Pool.Exec(ctx, `UPDATE games SET status = $1, updated_at = now() WHERE id = $2`, status, gameID)
	return err
}

func (db *DB) SetGameFinished(ctx context.Context, gameID uuid.UUID, winnerID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE games SET status = $1, win_condition_met = true, winner_id = $2, updated_at = now() WHERE id = $3`,
		models.GameStatusFinished, winnerID, gameID,
	)
	return err
}

func (db *DB) UpdatePlayerWolfForm(ctx context.Context, gameID, playerID uuid.UUID, isWolfForm bool) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE game_players SET is_wolf_form = $1 WHERE game_id = $2 AND player_id = $3`,
		isWolfForm, gameID, playerID,
	)
	return err
}

func (db *DB) GetPlayerByID(ctx context.Context, gameID, playerID uuid.UUID) (*models.GamePlayer, error) {
	var gp models.GamePlayer
	err := db.Pool.QueryRow(ctx,
		`SELECT game_id, player_id, slot_index, is_game_master, life, coins, gems, hero_id, declared_hero_id, is_wolf_form, is_eliminated, joined_at
		 FROM game_players WHERE game_id = $1 AND player_id = $2`,
		gameID, playerID,
	).Scan(&gp.GameID, &gp.PlayerID, &gp.SlotIndex, &gp.IsGameMaster, &gp.Life, &gp.Coins, &gp.Gems, &gp.HeroID, &gp.DeclaredHeroID, &gp.IsWolfForm, &gp.IsEliminated, &gp.JoinedAt)
	if err != nil {
		return nil, err
	}
	return &gp, nil
}
