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
// spectator == true means caller only watches if game already active
func (s *SessionService) JoinSession(
	sessionID, playerID string, spectator bool,
) (models.GameSession, error) {

	session, err := s.pb.FetchSession(sessionID)
	if err != nil {
		return session, err
	}

	// Already playing?
	inPlayers := contains(session.PlayerIDs, playerID)
	inSpecs := contains(session.State.SpectatorIDs, playerID)

	if session.IsActive {
		if inPlayers {
			return session, nil // already a player
		}
		if spectator {
			if inSpecs {
				return session, nil // already spectator
			}
			session.State.SpectatorIDs = append(session.State.SpectatorIDs, playerID)
			err = s.pb.UpdateSession(session.ID, session.State, session.IsActive)
			return session, err
		}
		return session, errors.New("game started; join as spectator")
	}

	// Lobby phase
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
	err = s.pb.UpdateSessionPlayers(session.ID, session.PlayerIDs)
	return session, err
}

func contains(list []string, v string) bool {
	for _, id := range list {
		if id == v {
			return true
		}
	}
	return false
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

	if len(session.State.ReadyIDs) != len(session.PlayerIDs) {
		return session, errors.New("all players must be ready")
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

func (s *SessionService) ToggleReady(
	sessionID, playerID string,
) (models.GameSession, error) {

	session, err := s.pb.FetchSession(sessionID)
	if err != nil {
		return session, err
	}

	// only before start and must be a listed player
	if session.IsActive || !contains(session.PlayerIDs, playerID) {
		return session, errors.New("cannot ready at this time")
	}

	ready := session.State.ReadyIDs
	if contains(ready, playerID) {
		// unready
		session.State.ReadyIDs = filter(ready, playerID)
	} else {
		session.State.ReadyIDs = append(ready, playerID)
	}
	err = s.pb.UpdateSession(session.ID, session.State, session.IsActive)
	return session, err
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
