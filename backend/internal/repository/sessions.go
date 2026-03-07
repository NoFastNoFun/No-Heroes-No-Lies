package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/models"
)

func (db *DB) CreatePlayer(ctx context.Context, username string) (*models.Player, error) {
	var p models.Player
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO players (id, username) VALUES (gen_random_uuid(), $1)
		 RETURNING id, username, created_at`,
		username,
	).Scan(&p.ID, &p.Username, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) CreateSession(ctx context.Context, playerID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO sessions (player_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		playerID, tokenHash, expiresAt,
	)
	return err
}

func (db *DB) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	var s models.Session
	err := db.Pool.QueryRow(ctx,
		`SELECT id, player_id, token_hash, expires_at, created_at
		 FROM sessions WHERE token_hash = $1`,
		tokenHash,
	).Scan(&s.ID, &s.PlayerID, &s.TokenHash, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) GetSessionByPlayerID(ctx context.Context, playerID uuid.UUID) (*models.Session, error) {
	var s models.Session
	err := db.Pool.QueryRow(ctx,
		`SELECT id, player_id, token_hash, expires_at, created_at
		 FROM sessions WHERE player_id = $1 AND expires_at > now()
		 ORDER BY expires_at DESC LIMIT 1`,
		playerID,
	).Scan(&s.ID, &s.PlayerID, &s.TokenHash, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) ExtendSessionExpiry(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE sessions SET expires_at = $1 WHERE id = $2`,
		expiresAt, sessionID,
	)
	return err
}
