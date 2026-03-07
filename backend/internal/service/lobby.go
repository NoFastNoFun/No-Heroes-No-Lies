package service

import (
	"context"
	"math/rand"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/gameloop"
	"github.com/no-heroes-no-lies/backend/internal/models"
	"github.com/no-heroes-no-lies/backend/internal/repository"
)

type LobbyService struct {
	db       *repository.DB
	registry *gameloop.Registry
}

func NewLobbyService(db *repository.DB, registry *gameloop.Registry) *LobbyService {
	return &LobbyService{db: db, registry: registry}
}

func (s *LobbyService) CreateGame(ctx context.Context, creatorID uuid.UUID, maxPlayers int) (*models.Game, error) {
	if maxPlayers < 2 || maxPlayers > 8 {
		return nil, ErrInvalidInput
	}
	return s.db.CreateGame(ctx, creatorID, maxPlayers)
}

func (s *LobbyService) ListGames(ctx context.Context, status models.GameStatus) ([]models.GameListItem, error) {
	return s.db.ListGamesByStatus(ctx, status)
}

func (s *LobbyService) GetGame(ctx context.Context, gameID uuid.UUID) (*models.Game, error) {
	return s.db.GetGameByID(ctx, gameID)
}

func (s *LobbyService) GetGamePlayers(ctx context.Context, gameID uuid.UUID) ([]models.GamePlayer, error) {
	return s.db.GetGamePlayers(ctx, gameID)
}

func (s *LobbyService) JoinGame(ctx context.Context, gameID, playerID uuid.UUID) (slotIndex int, err error) {
	game, err := s.db.GetGameByID(ctx, gameID)
	if err != nil {
		return 0, err
	}
	if game.Status != models.GameStatusLobby {
		return 0, ErrBadRequest
	}
	count, err := s.db.GamePlayerCount(ctx, gameID)
	if err != nil {
		return 0, err
	}
	if count >= game.MaxPlayers {
		return 0, ErrConflict
	}
	inGame, err := s.db.IsPlayerInGame(ctx, gameID, playerID)
	if err != nil {
		return 0, err
	}
	if inGame {
		return 0, ErrConflict
	}
	return s.db.JoinGame(ctx, gameID, playerID)
}

func (s *LobbyService) StartGame(ctx context.Context, gameID, playerID uuid.UUID) error {
	game, err := s.db.GetGameByID(ctx, gameID)
	if err != nil {
		return err
	}
	if game.Status != models.GameStatusLobby {
		return ErrBadRequest
	}
	isGM, err := s.db.IsGameMaster(ctx, gameID, playerID)
	if err != nil {
		return err
	}
	if !isGM {
		return ErrForbidden
	}
	count, err := s.db.GamePlayerCount(ctx, gameID)
	if err != nil {
		return err
	}
	if count != game.MaxPlayers {
		return ErrBadRequest
	}
	players, err := s.db.GetGamePlayers(ctx, gameID)
	if err != nil {
		return err
	}
	heroes, err := s.db.GetAllHeroes(ctx)
	if err != nil {
		return err
	}
	heroDeck := buildHeroDeck(heroes)
	activeMonsters, err := s.drawTwoActiveMonsters(ctx)
	if err != nil {
		return err
	}
	heroDeck = heroDeck[1:]
	playerHeroes := make(map[uuid.UUID]string)
	for i, p := range players {
		playerHeroes[p.PlayerID] = heroDeck[i]
	}
	heroDeck = heroDeck[len(players):]
	firstIndex := rand.Intn(len(players))
	firstPlayerID := players[firstIndex].PlayerID
	err = s.db.RunGameSetup(ctx, gameID, firstPlayerID, heroDeck, activeMonsters, playerHeroes)
	if err != nil {
		return err
	}
	if s.registry != nil {
		s.registry.Start(ctx, gameID)
	}
	return nil
}

func buildHeroDeck(heroes []repository.HeroRow) []string {
	var deck []string
	for _, h := range heroes {
		for i := 0; i < h.DefaultSpawnAmount; i++ {
			deck = append(deck, h.ID)
		}
	}
	randShuffleStrings(deck)
	return deck
}

func buildMonsterDecks(monsters []repository.MonsterRow) (deck []string, active [2]string) {
	var weighted []string
	for _, m := range monsters {
		for i := 0; i < m.SpawnWeight; i++ {
			weighted = append(weighted, m.ID)
		}
	}
	randShuffleStrings(weighted)
	if len(weighted) < 2 {
		return weighted, active
	}
	active[0] = weighted[0]
	active[1] = weighted[1]
	return weighted[2:], active
}

// drawTwoActiveMonsters draws two monsters by spawn weight (infinite deck).
func (s *LobbyService) drawTwoActiveMonsters(ctx context.Context) ([2]string, error) {
	var out [2]string
	for i := 0; i < 2; i++ {
		id, err := s.db.DrawRandomMonsterByWeight(ctx)
		if err != nil {
			return out, err
		}
		out[i] = id
	}
	return out, nil
}

func randShuffleStrings(s []string) {
	rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}
