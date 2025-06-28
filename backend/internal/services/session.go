package services

import (
	"errors"
	"math/rand"
	"time"

	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/pb"
)

// SessionService provides create / join / start helpers.
type SessionService struct {
	pb *pb.Client
}

// NewSessionService constructs a SessionService.
func NewSessionService(pbClient *pb.Client) *SessionService {
	return &SessionService{pb: pbClient}
}

// CreateSession creates a new session with the creator as first player.
func (s *SessionService) CreateSession(creatorID string) (models.GameSession, error) {
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
		PlayerIDs: []string{creatorID},
		State:     initialState,
		IsActive:  false,
	}
	return s.pb.InsertSession(session)
}

// JoinSession adds a player if the session is not active and below capacity.
func (s *SessionService) JoinSession(sessionID, playerID string) (models.GameSession, error) {
	session, err := s.pb.FetchSession(sessionID)
	if err != nil {
		return session, err
	}
	if session.IsActive {
		return session, errors.New("session already started")
	}
	for _, id := range session.PlayerIDs {
		if id == playerID {
			return session, nil // already joined
		}
	}
	if len(session.PlayerIDs) >= 15 {
		return session, errors.New("session full")
	}

	session.PlayerIDs = append(session.PlayerIDs, playerID)
	if err := s.pb.UpdateSessionPlayers(session.ID, session.PlayerIDs); err != nil {
		return session, err
	}
	return session, nil
}

// StartSession builds decks, deals cards, assigns life, and activates the game.
func (s *SessionService) StartSession(sessionID string) (models.GameSession, error) {
	session, err := s.pb.FetchSession(sessionID)
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

	// Build decks
	cards, err := s.pb.ListCards()
	if err != nil {
		return session, err
	}
	factor := (pCount + 4) / 5 // 1 for ≤5, 2 for 6-10, 3 for 11-15

	var heroDeck, monsterDeck []string
	for _, c := range cards {
		count := c.DefaultAmountPerSession * factor
		for i := 0; i < count; i++ {
			switch c.Type {
			case "hero":
				heroDeck = append(heroDeck, c.ID)
			case "monster":
				monsterDeck = append(monsterDeck, c.ID)
			}
		}
	}

	// Shuffle decks with a deterministic seed (stored for audit/debug)
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(heroDeck), func(i, j int) { heroDeck[i], heroDeck[j] = heroDeck[j], heroDeck[i] })
	rng.Shuffle(len(monsterDeck), func(i, j int) { monsterDeck[i], monsterDeck[j] = monsterDeck[j], monsterDeck[i] })

	// Burn 1 hero card per player to prevent counting
	burnCount := (pCount + 1) / 2  // half players, rounded up
	required := burnCount + pCount // cards needed after burn for dealing
	if len(heroDeck) < required {
		return session, errors.New("not enough hero cards")
	}
	heroDeck = heroDeck[burnCount:] // discard burned cards

	// ── DEAL hero cards ──
	var playerStates []models.PlayerState
	for _, id := range session.PlayerIDs {
		card := heroDeck[0]
		heroDeck = heroDeck[1:]
		playerStates = append(playerStates, models.PlayerState{
			ID:           id,
			CurrentHero:  card,
			CurrentAlibi: "",
			Life:         lifeForPlayers(pCount),
			Coins:        0,
			Gems:         0,
			Glory:        0,
			Hand:         []string{},
		})
	}

	// Draw 2 active monsters
	if len(monsterDeck) < 2 {
		return session, errors.New("not enough monster cards")
	}
	active := monsterDeck[:2]
	monsterDeck = monsterDeck[2:]

	// Randomise play order
	rng.Shuffle(len(session.PlayerIDs), func(i, j int) {
		session.PlayerIDs[i], session.PlayerIDs[j] = session.PlayerIDs[j], session.PlayerIDs[i]
	})

	// Update session state
	session.State.Players = playerStates
	session.State.TurnOrder = session.PlayerIDs
	session.State.CurrentTurn = session.PlayerIDs[0]
	session.State.HeroDeck = heroDeck
	session.State.MonsterDeck = monsterDeck
	session.State.ActiveMonsters = active
	session.State.Seed = seed
	session.IsActive = true

	if err := s.pb.UpdateSession(session.ID, session.State, session.IsActive); err != nil {
		return session, err
	}
	return session, nil
}

func lifeForPlayers(n int) int {
	switch {
	case n <= 5:
		return 2
	case n <= 10:
		return 3
	default:
		return 4
	}
}
