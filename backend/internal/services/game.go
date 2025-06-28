package services

import (
	"errors"

	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/pb"
)

// GameService provides game rule operations.
type GameService struct {
	pbClient *pb.Client
}

// NewGameService constructs a GameService.
func NewGameService(client *pb.Client) *GameService {
	return &GameService{pbClient: client}
}

// FetchSession returns a session by ID.
func (s *GameService) FetchSession(id string) (models.GameSession, error) {
	return s.pbClient.FetchSession(id)
}

// ApplyMove validates and applies a move.
func (s *GameService) ApplyMove(
	sessionID string,
	power models.Power,
	playerID string,
) error {
	session, err := s.FetchSession(sessionID)
	if err != nil {
		return err
	}

	if session.State.CurrentTurn != playerID {
		return errors.New("not your turn")
	}

	// TODO: full rule implementation

	// Advance turn
	session.State.CurrentTurn = nextPlayer(session.State.TurnOrder, playerID)

	// Update both state and is_active flag (unchanged here)
	return s.pbClient.UpdateSession(sessionID, session.State, session.IsActive)
}

func nextPlayer(order []string, current string) string {
	for i, id := range order {
		if id == current {
			return order[(i+1)%len(order)]
		}
	}
	return current
}
