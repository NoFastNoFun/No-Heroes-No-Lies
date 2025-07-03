package services

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"no-heroes-no-lies/internal/db"
	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/powers"
	"no-heroes-no-lies/internal/triggers"
)

// GameService provides game-rule operations.
type GameService struct {
	sessions *SessionService
}

// ChallengeWindowDuration is the duration for which a move can be challenged.
var ChallengeWindowDuration = 10 * time.Second // TODO: make configurable for difficulty options

func NewGameService(sessions *SessionService) *GameService {
	return &GameService{sessions: sessions}
}

// ── helpers ────────────────────────────────────────────────────────────────

func nextPlayer(order []string, current string) string {
	for i, id := range order {
		if id == current {
			return order[(i+1)%len(order)]
		}
	}
	return current // fallback (should never hit)
}

func heroStrength(
	_ *models.GameSession,
	player *models.PlayerState,
	card models.Card, // caller's cached hero card
) int {
	if player.BonusStrength > 0 {
		return player.BonusStrength
	}
	return card.Strength
}

// ── public API ─────────────────────────────────────────────────────────────

// FetchSession returns a session from Redis.
func (s *GameService) FetchSession(id string) (models.GameSession, error) {
	return s.sessions.FetchSession(id)
}

// ApplyMove = validate → mutate state → persist → log.
func (s *GameService) ApplyMove(
	sessionID string,
	playerID string,
	payload models.MovePayload,
) error {
	if s.isArchived(sessionID) {
		return errors.New("game is archived and cannot be modified")
	}
	session, err := s.FetchSession(sessionID)
	if err != nil {
		return err
	}
	if contains(session.State.SpectatorIDs, playerID) {
		return errors.New("spectators cannot act")
	}
	if session.State.CurrentTurn != playerID {
		return errors.New("not your turn")
	}

	// Reset order-0 power usage at the start of the turn
	session.State.Order0UsedBy = map[string]bool{}

	// Remove any lingering challenge window at the start of the turn
	session.State.LastMove = nil

	switch payload.Type {
	case "demask":
		if err := s.handleDemask(&session, playerID, payload); err != nil {
			return err
		}
	case "fight":
		if err := s.handleFight(&session, playerID, payload); err != nil {
			return err
		}
		// Start challenge window at alibi declaration, lasting until end of turn + ChallengeWindowDuration
		session.State.LastMove = &models.LastMove{
			ActorID:       playerID,
			Type:          "fight",
			DeclaredAlibi: payload.DeclaredAlibi,
			MonsterID:     payload.MonsterID,
			ExpiresAt:     time.Now().Add(ChallengeWindowDuration).UnixMilli(),
		}
	case "power":
		// Enforce power sequence: must pick a card, discard a card, declare alibi, use power
		if payload.DrawnCardID == "" || payload.DiscardCardID == "" || payload.DeclaredAlibi == "" || payload.PowerID == "" {
			return errors.New("power move must include drawn_card_id, discard_card_id, declared_alibi, and power_id")
		}
		if err := s.handlePower(&session, playerID, payload); err != nil {
			return err
		}
		// Start challenge window at alibi declaration, lasting until end of turn + ChallengeWindowDuration
		session.State.LastMove = &models.LastMove{
			ActorID:       playerID,
			Type:          "power",
			DeclaredAlibi: payload.DeclaredAlibi,
			ExpiresAt:     time.Now().Add(ChallengeWindowDuration).UnixMilli(),
		}
	default:
		return errors.New("unknown move type")
	}

	// advance turn
	session.State.CurrentTurn = nextPlayer(session.State.TurnOrder, playerID)

	// detect game outcome
	s.detectOutcome(&session.State)

	// persist
	if err := s.sessions.UpdateSession(session); err != nil {
		return err
	}

	// log move
	b, _ := json.Marshal(payload)
	return db.InsertMove(models.Move{
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
		// Demask successful - target loses life and must change hero
		target.Life--

		// Force target to draw a new hero card
		if len(session.State.HeroDeck) > 0 {
			newHero := session.State.HeroDeck[0]
			session.State.HeroDeck = session.State.HeroDeck[1:]
			s.changePlayerHero(session, target, newHero)
		} else {
			// If no cards in deck, try to recycle discard pile
			s.recycleDiscardPile(session)

			if len(session.State.HeroDeck) > 0 {
				// Successfully recycled, draw new hero
				newHero := session.State.HeroDeck[0]
				session.State.HeroDeck = session.State.HeroDeck[1:]
				s.changePlayerHero(session, target, newHero)
			} else {
				// No cards available at all, discard current hero and target has no hero
				if target.CurrentHero != "" {
					s.discardHeroCard(session, target.CurrentHero)
				}
				target.CurrentHero = ""
			}
		}
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

	heroCard, err := db.GetCard(p.DeclaredAlibi)
	if err != nil {
		return err
	}
	monsterCard, err := db.GetCard(p.MonsterID)
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

	heroStr := heroStrength(session, actor, heroCard)
	heroLost := monsterCard.Strength > heroStr
	var coinsDelta, gemsDelta int

	if heroLost {
		actor.Life--
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

	// After first monster resolution, process second if session.State.DualAttack:
	if session.State.DualAttack && len(session.State.ActiveMonsters) > 0 {
		session.State.DualAttack = false
		p.MonsterID = session.State.ActiveMonsters[0]
		// recursive one-more fight (reuse same logic)
		return s.handleFight(session, playerID, p)
	}
	return nil
}

func (s *GameService) handlePower(
	session *models.GameSession,
	playerID string,
	p models.MovePayload,
) error {
	// fetch power definition
	pwr, err := db.GetPower(p.PowerID)
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
	if err := eff(session, actor, p); err != nil {
		return err
	}

	// mark order 0 usage
	if pwr.Order == 0 {
		if session.State.Order0UsedBy == nil {
			session.State.Order0UsedBy = map[string]bool{}
		}
		session.State.Order0UsedBy[playerID] = true
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

	// detect game outcome
	s.detectOutcome(&session.State)

	return s.sessions.UpdateSession(session)
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
	_ *models.GameSession,
	player *models.PlayerState,
) ([]string, error) {
	if player.CurrentAlibi == "" {
		return nil, nil // no alibi declared ⇒ no passive
	}
	card, err := db.GetCard(player.CurrentAlibi)
	if err != nil {
		return nil, err
	}
	var acts []string
	for _, id := range card.PowerIDs {
		pwr, err := db.GetPower(id)
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
) bool {
	acts, err := alibiPassives(session, player)
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

// canStealGems checks if a target can prevent gem theft (used by powers)
func canStealGems(
	session *models.GameSession,
	tgt *models.PlayerState,
) bool {
	return !hasPassive(session, tgt, "keep_gems")
}

// giveGems adds gems to receiver and triggers "teamwork" passives.
func (s *GameService) giveGems(session *models.GameSession, recv *models.PlayerState, n int) {
	if n <= 0 {
		return
	}
	recv.Gems += n

	// event
	ev := triggers.Event{
		Type:    "gems_gained",
		ActorID: recv.ID,
		Amount:  n,
	}
	triggers.Dispatch(session, &ev)
}

// effectiveStrength calculates the effective strength of a player's declared alibi
func effectiveStrength(
	_ *models.GameSession,
	player *models.PlayerState,
) int {
	card, err := db.GetCard(player.CurrentAlibi)
	if err != nil {
		return 0
	}
	base := card.Strength
	if player.BonusStrength != 0 {
		base = player.BonusStrength
	}
	return base
}

func (s *GameService) detectOutcome(state *models.GameState) {
	if state.Draw || len(state.WinnerIDs) > 0 {
		// Archive the game if not already archived
		sessionID := ""
		if len(state.WinnerIDs) > 0 {
			sessionID = state.WinnerIDs[0] // fallback, but should use session.ID
		}
		// Ideally, pass session.ID from caller
		// For now, skip if sessionID is empty
		if sessionID != "" && !s.isArchived(sessionID) {
			session, err := s.sessions.FetchSession(sessionID)
			if err == nil {
				moves, _ := db.FetchMovesForSession(session.ID)
				_ = db.ArchiveGame(session, moves)
			}
		}
		return // already ended
	}

	var alive []*models.PlayerState
	for i := range state.Players {
		if state.Players[i].Life > 0 {
			alive = append(alive, &state.Players[i])
		}
	}

	switch {
	case len(alive) == 0: // everyone died same turn
		state.Draw = true
	case len(alive) == 1:
		state.WinnerIDs = []string{alive[0].ID}
	default:
		// coin victory check
		var coinWinners []string
		for _, p := range alive {
			if p.Coins >= len(state.Players)+1 {
				coinWinners = append(coinWinners, p.ID)
			}
		}
		if len(coinWinners) > 0 {
			state.WinnerIDs = coinWinners
		}
	}
}

// Forfeit marks a player as out, triggers win detection.
func (s *GameService) Forfeit(sessionID, playerID string) error {
	if s.isArchived(sessionID) {
		return errors.New("game is archived and cannot be modified")
	}
	session, err := s.FetchSession(sessionID)
	if err != nil {
		return err
	}
	// find player
	var pl *models.PlayerState
	for i := range session.State.Players {
		if session.State.Players[i].ID == playerID {
			pl = &session.State.Players[i]
			break
		}
	}
	if pl == nil {
		return errors.New("not in this game")
	}
	if pl.Life == 0 {
		return errors.New("already out")
	}

	// mark dead, move to spectators
	pl.Life = 0
	session.State.SpectatorIDs = append(session.State.SpectatorIDs, playerID)

	// outcome check
	s.detectOutcome(&session.State)
	if session.State.Draw || len(session.State.WinnerIDs) > 0 {
		session.IsActive = false
	}

	// persist
	if err := s.sessions.UpdateSession(session); err != nil {
		return err
	}

	// log move
	move, _ := json.Marshal(map[string]string{"type": "forfeit"})
	return db.InsertMove(models.Move{
		SessionID: session.ID,
		PlayerID:  playerID,
		Type:      "forfeit",
		Data:      string(move),
	})
}

// Helper function to discard a hero card (adds to discard pile and updates public discard)
func (s *GameService) discardHeroCard(session *models.GameSession, cardID string) {
	if cardID == "" {
		return
	}
	// Always keep at least one card in the discard pile unless burning
	if len(session.State.DiscardPile) == 0 {
		session.State.DiscardPile = append(session.State.DiscardPile, cardID)
	} else {
		// If burning, handle separately (not implemented here)
		session.State.DiscardPile = append(session.State.DiscardPile, cardID)
	}
	session.State.PublicDiscard = cardID
}

// Helper function to change a player's hero card
func (s *GameService) changePlayerHero(session *models.GameSession, player *models.PlayerState, newHeroID string) {
	// Discard current hero if exists
	if player.CurrentHero != "" {
		s.discardHeroCard(session, player.CurrentHero)
	}

	// Set new hero
	player.CurrentHero = newHeroID
}

// Helper function to recycle discard pile when hero deck is empty
func (s *GameService) recycleDiscardPile(session *models.GameSession) {
	if len(session.State.DiscardPile) <= 1 {
		return // always leave at least one card in the discard pile
	}
	// Move all but the last discarded card back to the deck
	cardsToRecycle := session.State.DiscardPile[:len(session.State.DiscardPile)-1]
	session.State.HeroDeck = append(session.State.HeroDeck, cardsToRecycle...)
	session.State.DiscardPile = session.State.DiscardPile[len(session.State.DiscardPile)-1:]
	session.State.PublicDiscard = session.State.DiscardPile[0] // last discarded card remains public
	// Shuffle the recycled deck
	rand.Shuffle(len(session.State.HeroDeck), func(i, j int) {
		session.State.HeroDeck[i], session.State.HeroDeck[j] = session.State.HeroDeck[j], session.State.HeroDeck[i]
	})
}

// Add a helper to check if a game is archived
func (s *GameService) isArchived(sessionID string) bool {
	// Try to fetch from archived_games; if found, it's archived
	_, _, err := db.GetArchivedGame(sessionID)
	return err == nil
}
