package services

import (
	"encoding/json"
	"errors"
	"time"

	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/pb"
	"no-heroes-no-lies/internal/powers"
)

// GameService provides game-rule operations.
type GameService struct {
	pbClient *pb.Client
}

func NewGameService(c *pb.Client) *GameService { return &GameService{pbClient: c} }

// ── helpers ────────────────────────────────────────────────────────────────

func nextPlayer(order []string, current string) string {
	for i, id := range order {
		if id == current {
			return order[(i+1)%len(order)]
		}
	}
	return current // fallback (should never hit)
}

// ── public API ─────────────────────────────────────────────────────────────

// FetchSession returns a PB record.
func (s *GameService) FetchSession(id string) (models.GameSession, error) {
	return s.pbClient.FetchSession(id)
}

// ApplyMove = validate → mutate state → persist → log.
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

	// persist
	if err := s.pbClient.UpdateSession(session.ID, session.State, session.IsActive); err != nil {
		return err
	}

	// log move
	b, _ := json.Marshal(payload)
	return s.pbClient.InsertMove(models.Move{
		SessionID: session.ID,
		PlayerID:  playerID,
		Type:      payload.Type,
		Data:      string(b),
	})
}

// ── move handlers ─────────────────────────────────────────────────────────

func (s *GameService) handleDemask(
	session *models.GameSession,
	playerID string,
	p models.MovePayload,
) error {
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
	// demask has no alibi declaration → no challenge window
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
	active := false
	for _, id := range session.State.ActiveMonsters {
		if id == p.MonsterID {
			active = true
			break
		}
	}
	if !active {
		return errors.New("monster not active")
	}

	heroCard, err := s.pbClient.GetCard(p.DeclaredAlibi)
	if err != nil {
		return err
	}
	monsterCard, err := s.pbClient.GetCard(p.MonsterID)
	if err != nil {
		return err
	}

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

	actor.CurrentAlibi = p.DeclaredAlibi

	heroLost := monsterCard.Strength > heroCard.Strength
	var coinsDelta, gemsDelta, lifeDelta int

	if heroLost {
		actor.Life--
		lifeDelta = -1
	} else {
		coinsDelta = monsterCard.Loot.Coins
		gemsDelta = monsterCard.Loot.Gems
		actor.Coins += coinsDelta
		s.giveGems(session, actor, gemsDelta)

		// remove slain monster
		var remain []string
		for _, id := range session.State.ActiveMonsters {
			if id != p.MonsterID {
				remain = append(remain, id)
			}
		}
		session.State.ActiveMonsters = remain

		// refill if needed
		if len(session.State.ActiveMonsters) < 2 && len(session.State.MonsterDeck) > 0 {
			next := session.State.MonsterDeck[0]
			session.State.MonsterDeck = session.State.MonsterDeck[1:]
			session.State.ActiveMonsters = append(session.State.ActiveMonsters, next)
		}
	}

	// 5-second challenge window
	session.State.LastMove = &models.LastMove{
		ActorID:       playerID,
		Type:          "fight",
		DeclaredAlibi: p.DeclaredAlibi,
		MonsterID:     p.MonsterID,
		CoinsDelta:    coinsDelta,
		GemsDelta:     gemsDelta,
		LifeDeltaID:   actor.ID,
		LifeDelta:     lifeDelta,
		ExpiresAt:     time.Now().Add(5 * time.Second).UnixMilli(),
	}
	return nil
}

func (s *GameService) handlePower(
	session *models.GameSession,
	playerID string,
	p models.MovePayload,
) error {
	// fetch power definition
	pwr, err := s.pbClient.GetPower(p.PowerID)
	if err != nil {
		return err
	}
	// order chain enforcement
	if pwr.Order == 1 {
		if session.State.Order0UsedBy == nil || !session.State.Order0UsedBy[playerID] {
			return errors.New("must use order-0 power first")
		}
	}
	// cost check
	actor := playerPtr(session, playerID)
	if actor.Gems < pwr.Cost {
		return errors.New("not enough gems")
	}
	actor.Gems -= pwr.Cost

	eff, ok := powers.Registry[pwr.Action]
	if !ok {
		return errors.New("power not implemented")
	}
	if err := eff(session, actor, p, s.pbClient); err != nil {
		return err
	}

	// mark order 0 usage
	if pwr.Order == 0 {
		if session.State.Order0UsedBy == nil {
			session.State.Order0UsedBy = map[string]bool{}
		}
		session.State.Order0UsedBy[playerID] = true
	}

	// 5-second challenge window
	session.State.LastMove = &models.LastMove{
		ActorID:       playerID,
		Type:          "power",
		DeclaredAlibi: p.DeclaredAlibi,
		ExpiresAt:     time.Now().Add(5 * time.Second).UnixMilli(),
	}
	return nil
}

// ── challenge resolution ──────────────────────────────────────────────────

func (s *GameService) ResolveChallenge(
	sessionID, challengerID string,
	nowMs int64,
) error {
	session, err := s.FetchSession(sessionID)
	if err != nil {
		return err
	}
	mv := session.State.LastMove
	if mv == nil || nowMs > mv.ExpiresAt {
		return errors.New("no challengable move")
	}
	if mv.ActorID == challengerID {
		return errors.New("cannot challenge yourself")
	}

	var liar, challenger *models.PlayerState
	for i := range session.State.Players {
		p := &session.State.Players[i]
		switch p.ID {
		case mv.ActorID:
			liar = p
		case challengerID:
			challenger = p
		}
	}
	if liar == nil || challenger == nil {
		return errors.New("players not found")
	}

	liarWasLying := liar.CurrentHero != mv.DeclaredAlibi
	if liarWasLying {
		// revert gains
		liar.Coins -= mv.CoinsDelta
		liar.Gems -= mv.GemsDelta
		// restore monster if needed
		if mv.Type == "fight" && mv.MonsterID != "" {
			session.State.ActiveMonsters = append(
				session.State.ActiveMonsters, mv.MonsterID,
			)
		}
		liar.Life--
		liar.CurrentHero = ""
	} else {
		challenger.Life--
	}

	session.State.LastMove = nil
	return s.pbClient.UpdateSession(session.ID, session.State, session.IsActive)
}

func playerPtr(s *models.GameSession, id string) *models.PlayerState {
	for i := range s.State.Players {
		if s.State.Players[i].ID == id {
			return &s.State.Players[i]
		}
	}
	return nil
}

func alibiPassives(
	session *models.GameSession,
	player *models.PlayerState,
	pbCli *pb.Client,
) ([]string, error) {
	if player.CurrentAlibi == "" {
		return nil, nil // no alibi declared ⇒ no passive
	}
	card, err := pbCli.GetCard(player.CurrentAlibi)
	if err != nil {
		return nil, err
	}
	var acts []string
	for _, id := range card.PowerIDs {
		pwr, err := pbCli.GetPower(id)
		if err != nil {
			return nil, err
		}
		if pwr.Type == "passive" {
			acts = append(acts, pwr.Action)
		}
	}
	return acts, nil
}

func hasPassive(
	session *models.GameSession,
	player *models.PlayerState,
	action string,
	pbCli *pb.Client,
) bool {
	acts, err := alibiPassives(session, player, pbCli)
	if err != nil {
		return false
	}
	for _, a := range acts {
		if a == action {
			return true
		}
	}
	return false
}

func (s *GameService) applyTurnStartPassives(session *models.GameSession) {
	p := playerPtr(session, session.State.CurrentTurn)
	if p == nil {
		return
	}

	// alternate_strength: flip 1 ↔ 7 each turn if the declared alibi has that passive
	if hasPassive(session, p, "alternate_strength", s.pbClient) {
		if p.BonusStrength == 1 {
			p.BonusStrength = 7
		} else {
			p.BonusStrength = 1
		}
	}
}

func canStealGems(
	session *models.GameSession,
	target *models.PlayerState,
	pbCli *pb.Client,
) bool {
	return !hasPassive(session, target, "keep_gems", pbCli)
}

// giveGems adds gems to receiver and triggers "teamwork" passives.
func (s *GameService) giveGems(
	session *models.GameSession,
	receiver *models.PlayerState,
	amount int,
) {
	if amount <= 0 {
		return
	}
	receiver.Gems += amount

	// Teamwork passive: every *other* player whose alibi hero has "teamwork"
	// also gains the same amount.
	for i := range session.State.Players {
		p := &session.State.Players[i]
		if p.ID == receiver.ID {
			continue
		}
		if hasPassive(session, p, "teamwork", s.pbClient) {
			p.Gems += amount
		}
	}
}
