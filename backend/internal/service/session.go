package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/auth"
	"github.com/no-heroes-no-lies/backend/internal/config"
	"github.com/no-heroes-no-lies/backend/internal/models"
	"github.com/no-heroes-no-lies/backend/internal/repository"
)

type SessionService struct {
	db     *repository.DB
	cfg    *config.Config
}

func NewSessionService(db *repository.DB, cfg *config.Config) *SessionService {
	return &SessionService{db: db, cfg: cfg}
}

type CreateSessionOutput struct {
	SessionToken string    `json:"session_token"`
	PlayerID     string    `json:"player_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *SessionService) CreateSession(ctx context.Context, username string) (*models.Player, *CreateSessionOutput, error) {
	if username == "" {
		return nil, nil, ErrInvalidInput
	}
	player, err := s.db.CreatePlayer(ctx, username)
	if err != nil {
		return nil, nil, err
	}
	ttl := s.cfg.SessionTTL
	token, err := auth.NewToken(player.ID, s.cfg.JWTSecret, ttl)
	if err != nil {
		return nil, nil, err
	}
	tokenHash := auth.HashToken(token)
	expiresAt := time.Now().Add(ttl)
	if err := s.db.CreateSession(ctx, player.ID, tokenHash, expiresAt); err != nil {
		return nil, nil, err
	}
	return player, &CreateSessionOutput{
		SessionToken: token,
		PlayerID:     player.ID.String(),
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *SessionService) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	claims, err := auth.ParseToken(token, s.cfg.JWTSecret)
	if err != nil {
		return uuid.Nil, ErrUnauthorized
	}
	playerID, err := uuid.Parse(claims.PlayerID)
	if err != nil {
		return uuid.Nil, ErrUnauthorized
	}
	tokenHash := auth.HashToken(token)
	sess, err := s.db.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return uuid.Nil, ErrUnauthorized
	}
	if time.Now().After(sess.ExpiresAt) {
		return uuid.Nil, ErrUnauthorized
	}
	if sess.PlayerID != playerID {
		return uuid.Nil, ErrUnauthorized
	}
	return playerID, nil
}

func (s *SessionService) ExtendSessionForPlayer(ctx context.Context, playerID uuid.UUID) error {
	sess, err := s.db.GetSessionByPlayerID(ctx, playerID)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(s.cfg.SessionTTL)
	return s.db.ExtendSessionExpiry(ctx, sess.ID, expiresAt)
}
