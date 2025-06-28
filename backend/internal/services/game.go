package services

import (
	"encoding/json"
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

// ApplyMove validates, mutates state, and logs.
func (s *GameService) ApplyMove(
	sessionID string,
	playerID string,
	payload models.MovePayload,
) error {
	session, err := s.FetchSession(sessionID)
	if err != nil {
		return err
	}
	if session.State.CurrentTurn != playerID {
		return errors.New("not your turn")
	}

	switch payload.Type {
	case "demask":
		if err := s.handleDemask(&session, playerID, payload); err != nil {
			return err
		}
	case "fight":
		if err := s.handleFight(&session, playerID, payload); err != nil {
			return err
		}
	case "power":
		if err := s.handlePower(&session, playerID, payload); err != nil {
			return err
		}
	default:
		return errors.New("unknown move type")
	}

	// advance turn
	session.State.CurrentTurn = nextPlayer(session.State.TurnOrder, playerID)

	if err := s.pbClient.UpdateSession(session.ID, session.State, session.IsActive); err != nil {
		return err
	}

	// log
	moveJSON, _ := json.Marshal(payload)
	return s.pbClient.InsertMove(models.Move{
		SessionID: session.ID,
		PlayerID:  playerID,
		Type:      payload.Type,
		Data:      string(moveJSON),
	})
}

func (s *GameService) handleDemask(
	session *models.GameSession,
	playerID string,
	p models.MovePayload,
) error {
	// locate players
	var actor, target *models.PlayerState
	for i := range session.State.Players {
		ptr := &session.State.Players[i]
		switch ptr.ID {
		case playerID:
			actor = ptr
		case p.TargetPlayer:
			target = ptr
		}
	}
	if actor == nil || target == nil {
		return errors.New("invalid target")
	}
	if actor.Gems < 6 {
		return errors.New("not enough gems")
	}
	actor.Gems -= 6
	if target.CurrentHero == p.GuessHero {
		target.Life--
		target.CurrentHero = ""
	}
	return nil
}

func (s *GameService) handleFight(
	session *models.GameSession,
	playerID string,
	p models.MovePayload,
) error {
	if p.MonsterID == "" || p.DeclaredAlibi == "" {
		return errors.New("monster_id and declared_alibi required")
	}

	// Ensure monster is currently attackable.
	found := false
	for _, mID := range session.State.ActiveMonsters {
		if mID == p.MonsterID {
			found = true
			break
		}
	}
	if !found {
		return errors.New("monster not active")
	}

	// Get card data.
	heroCard, err := s.pbClient.GetCard(p.DeclaredAlibi)
	if err != nil {
		return err
	}
	monsterCard, err := s.pbClient.GetCard(p.MonsterID)
	if err != nil {
		return err
	}

	// Locate acting player.
	var actor *models.PlayerState
	for i := range session.State.Players {
		if session.State.Players[i].ID == playerID {
			actor = &session.State.Players[i]
			break
		}
	}
	if actor == nil {
		return errors.New("player state not found")
	}

	// Store alibi (may equal true hero).
	actor.CurrentAlibi = p.DeclaredAlibi

	// Compare strengths.
	if monsterCard.Strength > heroCard.Strength {
		actor.Life--
	} else {
		// Player wins - grant loot.
		actor.Coins += monsterCard.Loot.Coins
		actor.Gems += monsterCard.Loot.Gems

		// Remove slain monster.
		var remaining []string
		for _, id := range session.State.ActiveMonsters {
			if id != p.MonsterID {
				remaining = append(remaining, id)
			}
		}
		session.State.ActiveMonsters = remaining

		// Draw next monster to keep two active if deck not empty.
		if len(session.State.ActiveMonsters) < 2 && len(session.State.MonsterDeck) > 0 {
			next := session.State.MonsterDeck[0]
			session.State.MonsterDeck = session.State.MonsterDeck[1:]
			session.State.ActiveMonsters = append(session.State.ActiveMonsters, next)
		}
	}

	// TODO: challenge-window (5 s liar call) not yet implemented.
	return nil
}

func (s *GameService) handlePower(
	session *models.GameSession,
	playerID string,
	raw interface{},
) error {
	// TODO: implement power resolution
	return nil
}

func nextPlayer(order []string, current string) string {
	for i, id := range order {
		if id == current {
			return order[(i+1)%len(order)]
		}
	}
	return current
}
