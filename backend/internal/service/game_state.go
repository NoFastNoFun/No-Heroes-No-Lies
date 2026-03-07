package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/models"
	"github.com/no-heroes-no-lies/backend/internal/repository"
)

type GameStateService struct {
	db *repository.DB
}

func NewGameStateService(db *repository.DB) *GameStateService {
	return &GameStateService{db: db}
}

func (s *GameStateService) GetState(ctx context.Context, gameID, playerID uuid.UUID) (*models.GameStateView, error) {
	game, err := s.db.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game.Status != models.GameStatusPlaying && game.Status != models.GameStatusFinished {
		return nil, ErrNotFound
	}
	inGame, err := s.db.IsPlayerInGame(ctx, gameID, playerID)
	if err != nil || !inGame {
		return nil, ErrNotFound
	}
	players, err := s.db.GetGamePlayers(ctx, gameID)
	if err != nil {
		return nil, err
	}
	turnState, err := s.db.GetGameTurnState(ctx, gameID)
	if err != nil {
		return nil, err
	}
	activeMonsters, err := s.db.GetActiveMonsters(ctx, gameID)
	if err != nil {
		return nil, err
	}
	deckSize, _ := s.db.GetHeroDeckSize(ctx, gameID)

	view := &models.GameStateView{
		GameID:         game.ID,
		Status:         game.Status,
		MaxPlayers:     game.MaxPlayers,
		RoundNumber:    game.RoundNumber,
		LastDiscardID:  game.LastDiscardHeroID,
		Phase:          turnState.Phase,
		ActingPlayerID: turnState.ActingPlayerID,
		ChallengeWindowEndsAt: turnState.ChallengeWindowEndsAt,
		ActiveMonsters: activeMonsters,
		DeckSize:       deckSize,
		WinnerID:       game.WinnerID,
	}
	for _, p := range players {
		view.Players = append(view.Players, models.PlayerStateView{
			PlayerID:       p.PlayerID,
			SlotIndex:      p.SlotIndex,
			Life:           p.Life,
			Coins:          p.Coins,
			Gems:           p.Gems,
			HeroID:         p.HeroID,
			DeclaredHeroID: p.DeclaredHeroID,
			IsEliminated:   p.IsEliminated,
		})
		if p.PlayerID == playerID {
			view.MySlot = p.SlotIndex
			view.MyHeroID = p.HeroID
			view.MyLife = p.Life
			view.MyCoins = p.Coins
			view.MyGems = p.Gems
		}
	}
	return view, nil
}
