package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"no-heroes-no-lies/internal/models"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

const sessionTTL = 24 * time.Hour

func InitRedis(addr string) {
	if addr == "" {
		addr = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
}

type SessionService struct{}

func NewSessionService() *SessionService {
	return &SessionService{}
}

func sessionKey(id string) string {
	return "session:" + id
}

func (s *SessionService) CreateSession(creatorID string) (models.GameSession, error) {
	id := fmt.Sprintf("sess_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
	initialState := models.GameState{
		Players:        []models.PlayerState{},
		TurnOrder:      []string{},
		CurrentTurn:    "",
		CoinPool:       0,
		GemPool:        0,
		HeroDeck:       []string{},
		MonsterDeck:    []string{},
		ActiveMonsters: []string{},
		Seed:           0,
	}
	session := models.GameSession{
		ID:        id,
		PlayerIDs: []string{creatorID},
		State:     initialState,
		IsActive:  false,
	}
	if err := saveSession(session); err != nil {
		return session, err
	}
	return session, nil
}

func (s *SessionService) FetchSession(id string) (models.GameSession, error) {
	return fetchSession(id)
}

func (s *SessionService) UpdateSession(session models.GameSession) error {
	return saveSession(session)
}

func (s *SessionService) JoinSession(sessionID, playerID string, spectator bool) (models.GameSession, error) {
	session, err := fetchSession(sessionID)
	if err != nil {
		return session, err
	}
	inPlayers := contains(session.PlayerIDs, playerID)
	inSpecs := contains(session.State.SpectatorIDs, playerID)

	if session.IsActive {
		if inPlayers {
			return session, nil
		}
		if spectator {
			if inSpecs {
				return session, nil
			}
			session.State.SpectatorIDs = append(session.State.SpectatorIDs, playerID)
			if err := saveSession(session); err != nil {
				return session, err
			}
			return session, nil
		}
		return session, errors.New("game started; join as spectator")
	}
	if spectator {
		return session, errors.New("can't spectate before start")
	}
	if inPlayers {
		return session, nil
	}
	if len(session.PlayerIDs) >= 15 {
		return session, errors.New("session full")
	}
	session.PlayerIDs = append(session.PlayerIDs, playerID)
	if err := saveSession(session); err != nil {
		return session, err
	}
	return session, nil
}

func (s *SessionService) StartSession(sessionID string) (models.GameSession, error) {
	session, err := fetchSession(sessionID)
	if err != nil {
		return session, err
	}
	if session.IsActive {
		return session, errors.New("already active")
	}
	pCount := len(session.PlayerIDs)
	if pCount < 2 {
		return session, errors.New("need at least 2 players")
	}
	if pCount > 15 {
		return session, errors.New("max 15 players")
	}
	if len(session.State.ReadyIDs) != len(session.PlayerIDs) {
		return session, errors.New("all players must be ready")
	}

	// Initialize random seed for this session if not already set
	if session.State.Seed == 0 {
		session.State.Seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(session.State.Seed))

	// Build hero and monster decks from static design data
	var heroDeck []string
	var monsterDeck []string
	for _, c := range design.AllCards() {
		switch c.Type {
		case "hero":
			for i := 0; i < c.DefaultAmountPerSession; i++ {
				heroDeck = append(heroDeck, c.ID)
			}
		case "monster":
			// Use DefaultAmountPerSession for monsters as well; if zero, default to 1 copy
			count := c.DefaultAmountPerSession
			if count <= 0 {
				count = 1
			}
			for i := 0; i < count; i++ {
				monsterDeck = append(monsterDeck, c.ID)
			}
		}
	}

	// Shuffle decks deterministically based on session seed
	rng.Shuffle(len(heroDeck), func(i, j int) {
		heroDeck[i], heroDeck[j] = heroDeck[j], heroDeck[i]
	})
	rng.Shuffle(len(monsterDeck), func(i, j int) {
		monsterDeck[i], monsterDeck[j] = monsterDeck[j], monsterDeck[i]
	})

	// Initialize active monsters (up to 2 at start)
	activeCount := 2
	if activeCount > len(monsterDeck) {
		activeCount = len(monsterDeck)
	}
	activeMonsters := append([]string{}, monsterDeck[:activeCount]...)
	monsterDeck = monsterDeck[activeCount:]

	session.State.HeroDeck = heroDeck
	session.State.MonsterDeck = monsterDeck
	session.State.ActiveMonsters = activeMonsters

	session.IsActive = true
	if err := saveSession(session); err != nil {
		return session, err
	}
	return session, nil
}

func (s *SessionService) ToggleReady(sessionID, playerID string) (models.GameSession, error) {
	session, err := fetchSession(sessionID)
	if err != nil {
		return session, err
	}
	if session.IsActive || !contains(session.PlayerIDs, playerID) {
		return session, errors.New("cannot ready at this time")
	}
	ready := session.State.ReadyIDs
	if contains(ready, playerID) {
		session.State.ReadyIDs = filter(ready, playerID)
	} else {
		session.State.ReadyIDs = append(ready, playerID)
	}
	if err := saveSession(session); err != nil {
		return session, err
	}
	return session, nil
}

// --- Helpers ---
func contains(list []string, v string) bool {
	for _, id := range list {
		if id == v {
			return true
		}
	}
	return false
}

func filter(arr []string, drop string) []string {
	var out []string
	for _, v := range arr {
		if v != drop {
			out = append(out, v)
		}
	}
	return out
}

func saveSession(session models.GameSession) error {
	ctx := context.Background()
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, sessionKey(session.ID), data, sessionTTL).Err()
}

func fetchSession(id string) (models.GameSession, error) {
	ctx := context.Background()
	val, err := redisClient.Get(ctx, sessionKey(id)).Result()
	if err != nil {
		return models.GameSession{}, err
	}
	var session models.GameSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return models.GameSession{}, err
	}
	return session, nil
}
