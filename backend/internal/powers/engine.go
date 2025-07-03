package powers

import (
	"errors"
	"math/rand"

	"no-heroes-no-lies/internal/db"
	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/triggers"
)

/*
   Effect signature: no pb.Client
   - *session  (mutated)
   - *actor    (player executing the power)
   - payload   (raw MovePayload: target ids, etc.)

   All functions must:
   1. respect actor.Gems >= cost (cost already deducted by caller)
   2. return error on invalid params
*/

// ------------------------------------------------------------------
// registry helpers
// ------------------------------------------------------------------

// type Effect func(*models.GameSession, *models.PlayerState, models.MovePayload) error

var Registry = map[string]func(*models.GameSession, *models.PlayerState, models.MovePayload) error{}

func register(action string, fn func(*models.GameSession, *models.PlayerState, models.MovePayload) error) {
	Registry[action] = fn
}

// util
func findPlayer(s *models.GameSession, id string) *models.PlayerState {
	for i := range s.State.Players {
		if s.State.Players[i].ID == id {
			return &s.State.Players[i]
		}
	}
	return nil
}

// ---- local helpers (powers pkg only) ---------------------

// fetch passive actions on a player's DECLARED alibi
func alibiPassives(
	s *models.GameSession,
	p *models.PlayerState,
) ([]string, error) {
	if p.CurrentAlibi == "" {
		return nil, nil
	}
	card, err := db.GetCard(p.CurrentAlibi)
	if err != nil {
		return nil, err
	}
	var acts []string
	for _, id := range card.PowerIDs {
		pw, err := db.GetPower(id)
		if err != nil {
			return nil, err
		}
		if pw.Type == "passive" {
			acts = append(acts, pw.Action)
		}
	}
	return acts, nil
}

func hasPassive(
	s *models.GameSession,
	p *models.PlayerState,
	act string,
) bool {
	acts, err := alibiPassives(s, p)
	if err != nil {
		return false
	}
	for _, a := range acts {
		if a == act {
			return true
		}
	}
	return false
}

// canStealGems is used by steal_gem powers to check if target can prevent theft
func canStealGems(
	s *models.GameSession,
	target *models.PlayerState,
) bool {
	return !hasPassive(s, target, "keep_gems")
}

// Helper: ensure hero deck is not empty, recycling discard pile if needed
func ensureHeroDeckNotEmpty(s *models.GameSession, rng *rand.Rand) error {
	if len(s.State.HeroDeck) == 0 {
		if len(s.State.DiscardPile) > 0 {
			s.State.HeroDeck = append(s.State.HeroDeck, s.State.DiscardPile...)
			s.State.DiscardPile = nil
			s.State.PublicDiscard = "" // clear public discard since we're recycling
			rng.Shuffle(len(s.State.HeroDeck), func(i, j int) {
				s.State.HeroDeck[i], s.State.HeroDeck[j] = s.State.HeroDeck[j], s.State.HeroDeck[i]
			})
		} else {
			return errors.New("hero deck empty and no cards to recycle")
		}
	}
	return nil
}

// Helper: get per-session rand.Rand seeded from session.State.Seed
func sessionRand(s *models.GameSession) *rand.Rand {
	return rand.New(rand.NewSource(s.State.Seed))
}

// ------------------------------------------------------------------
// ACTIVE POWERS (order 0 / 1)
// ------------------------------------------------------------------

// change_monster - swap one active monster with top of deck
func init() {
	register("change_monster", func(s *models.GameSession, _ *models.PlayerState, _ models.MovePayload) error {
		if len(s.State.MonsterDeck) == 0 || len(s.State.ActiveMonsters) == 0 {
			return errors.New("no monsters to swap")
		}
		idx := rand.Intn(len(s.State.ActiveMonsters))
		draw := s.State.MonsterDeck[0]
		s.State.MonsterDeck = s.State.MonsterDeck[1:]
		s.State.ActiveMonsters[idx] = draw
		return nil
	})
}

// fight_monster - handled elsewhere by "fight" move, power just flags free fight
func init() {
	register("fight_monster", func(_ *models.GameSession, _ *models.PlayerState, _ models.MovePayload) error {
		// ApplyMove already does a fight; cost 1 gem taken by caller
		return nil
	})
}

// fight_player_with_discarded_card - strength duel with discarded hero
func init() {
	register("fight_player_with_discarded_card",
		func(s *models.GameSession, a *models.PlayerState, p models.MovePayload) error {

			t := findPlayer(s, p.TargetPlayer)
			if t == nil {
				return errors.New("target not found")
			}
			if p.DiscardCardID == "" {
				return errors.New("discard_card_id missing")
			}
			card, err := db.GetCard(p.DiscardCardID)
			if err != nil {
				return err
			}
			att := card.Strength + a.BonusStrength
			a.BonusStrength = 0

			def := effectiveStrength(s, t)
			if att > def {
				t.Life--
			} else {
				a.Life--
			}
			return nil
		})
}

func effectiveStrength(s *models.GameSession, p *models.PlayerState) int {
	card, err := db.GetCard(p.CurrentAlibi)
	if err != nil {
		return 0
	}
	base := card.Strength
	if p.BonusStrength != 0 {
		base = p.BonusStrength
	}
	return base
}

// add_strength_to_challenger - +1 only for same-turn fight_player_with_discarded_card
func init() {
	register("add_strength_to_challenger", func(_ *models.GameSession, a *models.PlayerState, _ models.MovePayload) error {
		a.BonusStrength = 1
		return nil
	})
}

// shuffle_heroes_deck - put all non-burned discards back & shuffle
func init() {
	register("shuffle_heroes_deck", func(s *models.GameSession, _ *models.PlayerState, _ models.MovePayload) error {
		s.State.HeroDeck = append(s.State.HeroDeck, s.State.DiscardPile...) // need DiscardPile field
		s.State.DiscardPile = nil
		rand.Shuffle(len(s.State.HeroDeck), func(i, j int) { s.State.HeroDeck[i], s.State.HeroDeck[j] = s.State.HeroDeck[j], s.State.HeroDeck[i] })
		return nil
	})
}

// change_hero - draw new hero, discard current
func init() {
	register("change_hero", func(s *models.GameSession, a *models.PlayerState, _ models.MovePayload) error {
		rng := sessionRand(s)
		if err := ensureHeroDeckNotEmpty(s, rng); err != nil {
			return err
		}
		if a.CurrentHero != "" {
			s.State.DiscardPile = append(s.State.DiscardPile, a.CurrentHero)
			s.State.PublicDiscard = a.CurrentHero
		}
		a.CurrentHero = s.State.HeroDeck[0]
		s.State.HeroDeck = s.State.HeroDeck[1:]
		return nil
	})
}

// exchange_gems_with_player
func init() {
	register("exchange_gems_with_player", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target not found")
		}
		a.Gems, target.Gems = target.Gems, a.Gems
		return nil
	})
}

// exchange_2coins_for_life
func init() {
	register("exchange_2coins_for_life", func(_ *models.GameSession, a *models.PlayerState, _ models.MovePayload) error {
		if a.Coins < 2 {
			return errors.New("need 2 coins")
		}
		a.Coins -= 2
		a.Life++
		return nil
	})
}

// gain_2gems
func init() {
	register("gain_2gems", func(_ *models.GameSession, a *models.PlayerState, _ models.MovePayload) error {
		a.Gems += 2
		return nil
	})
}

// gain_life
func init() {
	register("gain_life", func(_ *models.GameSession, a *models.PlayerState, _ models.MovePayload) error {
		a.Life++
		return nil
	})
}

// see_player_card
func init() {
	register("see_player_card", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target not found")
		}
		// In real implementation send secret notification; here: no state change
		_ = target.CurrentHero
		return nil
	})
}

// remove_coin
func init() {
	register("remove_coin", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil || target.Coins == 0 {
			return errors.New("target has no coins")
		}
		target.Coins--
		return nil
	})
}

// shoot_player
func init() {
	register("shoot_player",
		func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload) error {

			t := findPlayer(s, p.TargetPlayer)
			if t == nil {
				return errors.New("target not found")
			}
			ok := (p.GuessHero != "" && p.GuessHero == t.CurrentHero)
			if ok {
				t.Life--
			}
			s.State.LastShootOK = ok
			return nil
		})

}

// all_in
func init() {
	register("all_in",
		func(s *models.GameSession, a *models.PlayerState, _ models.MovePayload) error {
			if !s.State.LastShootOK {
				a.Gems = 0
				return nil
			}
			a.Gems *= 2
			return nil
		})
}

// steal_gem order0
func init() {
	register("steal_gem_0", func(
		s *models.GameSession, a *models.PlayerState, p models.MovePayload) error {
		t := findPlayer(s, p.TargetPlayer)
		if t == nil || t.Gems == 0 {
			return errors.New("target empty")
		}
		ev := triggers.Event{
			Type:     "steal_attempt",
			ActorID:  a.ID,
			TargetID: t.ID,
			Amount:   1,
		}
		if triggers.Dispatch(s, &ev) { // keep_gems may cancel
			return errors.New("steal blocked")
		}

		// success
		t.Gems--
		a.Gems++
		// teamwork duplicate
		ev = triggers.Event{Type: "gems_gained", ActorID: a.ID, Amount: 1}
		triggers.Dispatch(s, &ev)
		return nil
	})
	register("steal_gem_1", Registry["steal_gem_0"])

}

// steal_gem order-1 just reuses the same effect
func init() {
	register("steal_gem_1", Registry["steal_gem_0"])
}

// parry_this_you_f_casual
func init() {
	register("parry_this_you_f_casual", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target not found")
		}
		target.Life--
		return nil
	})
}

// steal_coin
func init() {
	register("steal_coin", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil || target.Coins == 0 {
			return errors.New("target empty")
		}
		target.Coins--
		a.Coins++
		return nil
	})
}

// mimic_hero - Copy another player's hero abilities temporarily
func init() {
	register("mimic_hero", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target player not found")
		}
		a.CurrentAlibi = target.CurrentAlibi // Temporarily copy hero abilities
		// You may want to set a flag or timer to revert this after a turn
		return nil
	})
}

// mimic_power
func init() {
	register("mimic_power", func(_ *models.GameSession, _ *models.PlayerState, _ models.MovePayload) error {
		return errors.New("mimic_power not yet implemented")
	})
}

// dual_attack
func init() {
	register("dual_attack",
		func(s *models.GameSession, _ *models.PlayerState, _ models.MovePayload) error {
			s.State.DualAttack = true
			return nil
		})
}

// blind_draw
func init() {
	register("blind_draw", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload) error {
		rng := sessionRand(s)
		if err := ensureHeroDeckNotEmpty(s, rng); err != nil {
			return err
		}
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target not found")
		}
		s.State.DiscardPile = append(s.State.DiscardPile, target.CurrentHero)
		target.CurrentHero = s.State.HeroDeck[0]
		s.State.HeroDeck = s.State.HeroDeck[1:]
		target.CurrentAlibi = "" // must bluff again
		return nil
	})
}

// force_transform
func init() {
	register("force_transform", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload) error {
		rng := sessionRand(s)
		if err := ensureHeroDeckNotEmpty(s, rng); err != nil {
			return err
		}
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target not found")
		}
		s.State.DiscardPile = append(s.State.DiscardPile, target.CurrentHero)
		target.CurrentHero = s.State.HeroDeck[0]
		s.State.HeroDeck = s.State.HeroDeck[1:]
		target.CurrentAlibi = ""
		return nil
	})
}

// Helper: find players holding a specific card
func findPlayersWithCard(s *models.GameSession, cardID string) []*models.PlayerState {
	var players []*models.PlayerState
	for i := range s.State.Players {
		if s.State.Players[i].CurrentHero == cardID {
			players = append(players, &s.State.Players[i])
		}
	}
	return players
}

// Helper: remove a card from discard pile, return true if found and removed
func removeCardFromDiscardPile(s *models.GameSession, cardID string) bool {
	for i, card := range s.State.DiscardPile {
		if card == cardID {
			s.State.DiscardPile = append(s.State.DiscardPile[:i], s.State.DiscardPile[i+1:]...)
			return true
		}
	}
	return false
}

// Helper: remove a card from deck, return true if found and removed
func removeCardFromDeck(s *models.GameSession, cardID string) bool {
	for i, card := range s.State.HeroDeck {
		if card == cardID {
			s.State.HeroDeck = append(s.State.HeroDeck[:i], s.State.HeroDeck[i+1:]...)
			return true
		}
	}
	return false
}

// Helper: execute a card from a player (force them to draw new hero and lose life)
func executeCardFromPlayer(s *models.GameSession, player *models.PlayerState, rng *rand.Rand) error {
	if err := ensureHeroDeckNotEmpty(s, rng); err != nil {
		return err
	}

	s.State.DiscardPile = append(s.State.DiscardPile, player.CurrentHero)
	player.CurrentHero = s.State.HeroDeck[0]
	s.State.HeroDeck = s.State.HeroDeck[1:]
	player.CurrentAlibi = "" // must bluff again
	player.Life--            // lose a life point
	return nil
}

// Helper: execute a unique card (targets a player directly)
func executeUniqueCard(s *models.GameSession, cardID string, playersWithCard []*models.PlayerState, rng *rand.Rand) (bool, error) {
	if len(playersWithCard) == 0 {
		return false, nil
	}

	// Randomly choose one player if multiple have the same card
	chosenPlayer := playersWithCard[0]
	if len(playersWithCard) > 1 {
		chosenPlayer = playersWithCard[rng.Intn(len(playersWithCard))]
	}

	if err := executeCardFromPlayer(s, chosenPlayer, rng); err != nil {
		return false, err
	}

	return true, nil
}

// Helper: execute a non-unique card (prioritizes discard pile, then deck, then players)
func executeNonUniqueCard(s *models.GameSession, cardID string, playersWithCard []*models.PlayerState, rng *rand.Rand) (bool, error) {
	// Try discard pile first
	if removeCardFromDiscardPile(s, cardID) {
		return true, nil
	}

	// Try deck next
	if removeCardFromDeck(s, cardID) {
		return true, nil
	}

	// Last resort: target a random player with the card
	if len(playersWithCard) > 0 {
		chosenPlayer := playersWithCard[rng.Intn(len(playersWithCard))]
		if err := executeCardFromPlayer(s, chosenPlayer, rng); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// execution - burn a hero card permanently from the session
func init() {
	register("execution", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload) error {
		rng := sessionRand(s)
		if p.DiscardCardID == "" {
			return errors.New("discard_card_id required for execution")
		}

		// Get card info to check if it's unique
		card, err := db.GetCard(p.DiscardCardID)
		if err != nil {
			return err
		}

		isUnique := card.DefaultAmountPerSession == 1
		playersWithCard := findPlayersWithCard(s, p.DiscardCardID)

		// Execute the card based on uniqueness
		var cardExecuted bool
		var execErr error

		if isUnique && len(playersWithCard) > 0 {
			cardExecuted, execErr = executeUniqueCard(s, p.DiscardCardID, playersWithCard, rng)
		} else {
			cardExecuted, execErr = executeNonUniqueCard(s, p.DiscardCardID, playersWithCard, rng)
		}

		if execErr != nil {
			return execErr
		}

		// Clear public discard if it was the executed card
		if s.State.PublicDiscard == p.DiscardCardID {
			s.State.PublicDiscard = ""
		}

		// Add the card to burned pile only if it was actually executed
		if cardExecuted {
			s.State.BurnedCards = append(s.State.BurnedCards, p.DiscardCardID)
		}

		return nil
	})
}

// keep_gems passive → cancel steal_attempt
func init() {
	triggers.Register("steal_attempt", func(s *models.GameSession, ev *triggers.Event) {
		tgt := findPlayer(s, ev.TargetID)
		if tgt != nil && hasPassive(s, tgt, "keep_gems") {
			ev.Cancel = true
		}
	})
}

// Updated gems_gained trigger for teamwork and tribute passives
func init() {
	triggers.Register("gems_gained", func(s *models.GameSession, ev *triggers.Event) {
		// Find the Chevalier (teamwork passive)
		var chevalier *models.PlayerState
		for i := range s.State.Players {
			if hasPassive(s, &s.State.Players[i], "teamwork") {
				chevalier = &s.State.Players[i]
				break
			}
		}

		// Track all gainers for tribute pass
		gainers := map[string]int{}

		// 1. Apply teamwork: Chevalier copies the gain (if not the gainer)
		if chevalier != nil && ev.ActorID != chevalier.ID {
			chevalier.Gems += ev.Amount
			gainers[chevalier.ID] = ev.Amount
		}
		// The original gainer
		gainers[ev.ActorID] = ev.Amount

		// 2. Apply tribute: Mad Prince taxes all gainers (except himself)
		var prince *models.PlayerState
		for i := range s.State.Players {
			if hasPassive(s, &s.State.Players[i], "tribute") {
				prince = &s.State.Players[i]
				break
			}
		}
		if prince != nil {
			for gainerID, amount := range gainers {
				if gainerID == prince.ID {
					continue
				}
				tribute := amount / 2
				// Find the gainer
				for i := range s.State.Players {
					if s.State.Players[i].ID == gainerID {
						s.State.Players[i].Gems -= tribute
						if s.State.Players[i].Gems < 0 {
							s.State.Players[i].Gems = 0
						}
						break
					}
				}
				prince.Gems += tribute
			}
		}
	})
}

// command_the_dead - Use discarded hero to fight monster and gain loot
func init() {
	register("command_the_dead", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload) error {
		if p.DiscardCardID == "" || p.MonsterID == "" {
			return errors.New("missing discard_card_id or monster_id")
		}
		// Simulate a fight between the discarded hero and the monster
		card, err := db.GetCard(p.DiscardCardID)
		if err != nil {
			return err
		}
		monster, err := db.GetCard(p.MonsterID)
		if err != nil {
			return err
		}
		if card.Strength >= monster.Strength {
			// Award loot to the player
			a.Coins += monster.Loot.Coins
			a.Gems += monster.Loot.Gems
			// Optionally, remove monster from play
		}
		// No penalty if the fight fails
		return nil
	})
}
