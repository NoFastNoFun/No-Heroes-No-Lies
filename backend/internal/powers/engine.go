package powers

import (
	"errors"
	"math/rand"

	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/pb"
)

/*
   Effect signature:
   - *session  (mutated)
   - *actor    (player executing the power)
   - payload   (raw MovePayload: target ids, etc.)
   - pbCli     (for card / power look-ups)

   All functions must:
   1. respect actor.Gems >= cost (cost already deducted by caller)
   2. return error on invalid params
*/

// ------------------------------------------------------------------
// registry helpers
// ------------------------------------------------------------------

type Effect func(*models.GameSession, *models.PlayerState, models.MovePayload, *pb.Client) error

var Registry = map[string]Effect{}

func register(action string, fn Effect) { Registry[action] = fn }

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
	cli *pb.Client,
) ([]string, error) {
	if p.CurrentAlibi == "" {
		return nil, nil
	}
	card, err := cli.GetCard(p.CurrentAlibi)
	if err != nil {
		return nil, err
	}
	var acts []string
	for _, id := range card.PowerIDs {
		pw, err := cli.GetPower(id)
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
	cli *pb.Client,
) bool {
	acts, err := alibiPassives(s, p, cli)
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

func canStealGems(
	s *models.GameSession,
	target *models.PlayerState,
	cli *pb.Client,
) bool {
	return !hasPassive(s, target, "keep_gems", cli)
}

// ------------------------------------------------------------------
// ACTIVE POWERS (order 0 / 1)
// ------------------------------------------------------------------

// change_monster - swap one active monster with top of deck
func init() {
	register("change_monster", func(s *models.GameSession, _ *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
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
	register("fight_monster", func(_ *models.GameSession, _ *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
		// ApplyMove already does a fight; cost 1 gem taken by caller
		return nil
	})
}

// fight_player_with_discarded_card - strength duel with discarded hero
func init() {
	register("fight_player_with_discarded_card",
		func(s *models.GameSession, a *models.PlayerState, p models.MovePayload, cli *pb.Client) error {

			t := findPlayer(s, p.TargetPlayer)
			if t == nil {
				return errors.New("target not found")
			}
			if p.DiscardCardID == "" {
				return errors.New("discard_card_id missing")
			}
			card, err := cli.GetCard(p.DiscardCardID)
			if err != nil {
				return err
			}
			att := card.Strength + a.BonusStrength
			a.BonusStrength = 0

			def := effectiveStrength(s, t, cli)
			if att > def {
				t.Life--
			} else {
				a.Life--
			}
			return nil
		})
}

func targetBonusStrength(p *models.PlayerState) int { return 0 } // placeholder

// add_strength_to_challenger - +1 only for same-turn fight_player_with_discarded_card
func init() {
	register("add_strength_to_challenger", func(_ *models.GameSession, a *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
		a.BonusStrength = 1
		return nil
	})
}

// shuffle_heroes_deck - put all non-burned discards back & shuffle
func init() {
	register("shuffle_heroes_deck", func(s *models.GameSession, _ *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
		s.State.HeroDeck = append(s.State.HeroDeck, s.State.DiscardPile...) // need DiscardPile field
		s.State.DiscardPile = nil
		rand.Shuffle(len(s.State.HeroDeck), func(i, j int) { s.State.HeroDeck[i], s.State.HeroDeck[j] = s.State.HeroDeck[j], s.State.HeroDeck[i] })
		return nil
	})
}

// change_hero - draw new hero, discard current
func init() {
	register("change_hero", func(s *models.GameSession, a *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
		if len(s.State.HeroDeck) == 0 {
			return errors.New("hero deck empty")
		}
		s.State.DiscardPile = append(s.State.DiscardPile, a.CurrentHero)
		a.CurrentHero = s.State.HeroDeck[0]
		s.State.HeroDeck = s.State.HeroDeck[1:]
		return nil
	})
}

// exchange_gems_with_player
func init() {
	register("exchange_gems_with_player", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload, _ *pb.Client) error {
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
	register("exchange_2coins_for_life", func(_ *models.GameSession, a *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
		if a.Coins < 2 {
			return errors.New("need 2 coins")
		}
		a.Coins -= 2
		a.Life++
		return nil
	})
}

// gain_2gems already registered above

// gain_life
func init() {
	register("gain_life", func(_ *models.GameSession, a *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
		a.Life++
		return nil
	})
}

// see_player_card
func init() {
	register("see_player_card", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload, _ *pb.Client) error {
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
	register("remove_coin", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload, _ *pb.Client) error {
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
		func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload, _ *pb.Client) error {

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
		func(s *models.GameSession, a *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
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
		s *models.GameSession,
		a *models.PlayerState,
		p models.MovePayload,
		cli *pb.Client,
	) error {
		t := findPlayer(s, p.TargetPlayer)
		if t == nil || t.Gems == 0 {
			return errors.New("target empty")
		}
		if !canStealGems(s, t, cli) {
			return errors.New("protected")
		}

		// steal 1 gem
		t.Gems--
		a.Gems++

		// TEAMWORK passive duplicates gain to other players
		for i := range s.State.Players {
			pl := &s.State.Players[i]
			if pl.ID == a.ID {
				continue
			}
			if hasPassive(s, pl, "teamwork", cli) {
				pl.Gems++
			}
		}
		return nil
	})
}

// steal_gem order-1 just reuses the same effect
func init() {
	register("steal_gem_1", Registry["steal_gem_0"])
}

// parry_this_you_f_casual
func init() {
	register("parry_this_you_f_casual", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload, _ *pb.Client) error {
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
	register("steal_coin", func(s *models.GameSession, a *models.PlayerState, p models.MovePayload, _ *pb.Client) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil || target.Coins == 0 {
			return errors.New("target empty")
		}
		target.Coins--
		a.Coins++
		return nil
	})
}

// mimic_hero
func init() {
	register("mimic_power",
		func(s *models.GameSession, a *models.PlayerState, p models.MovePayload, cli *pb.Client) error {
			t := findPlayer(s, p.TargetPlayer)
			if t == nil {
				return errors.New("target?")
			}
			card, err := cli.GetCard(t.CurrentHero)
			if err != nil {
				return err
			}
			if len(card.PowerIDs) == 0 {
				return errors.New("target hero has no powers")
			}
			pw, err := cli.GetPower(card.PowerIDs[0]) // order-0
			if err != nil {
				return err
			}
			if pw.Type == "passive" {
				return errors.New("target power is passive")
			}
			eff, ok := Registry[pw.Action]
			if !ok {
				return errors.New("copied power unimplemented")
			}
			return eff(s, a, p, cli) // reuse same payload
		})
}

// mimic_power
func init() {
	register("mimic_power", func(_ *models.GameSession, _ *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
		return errors.New("mimic_power not yet implemented")
	})
}

// dual_attack
func init() {
	register("dual_attack",
		func(s *models.GameSession, _ *models.PlayerState, _ models.MovePayload, _ *pb.Client) error {
			s.State.DualAttack = true
			return nil
		})
}

// blind_draw
func init() {
	register("blind_draw", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload, _ *pb.Client) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target not found")
		}
		if len(s.State.HeroDeck) == 0 {
			return errors.New("deck empty")
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
	register("force_transform", func(s *models.GameSession, _ *models.PlayerState, p models.MovePayload, _ *pb.Client) error {
		target := findPlayer(s, p.TargetPlayer)
		if target == nil {
			return errors.New("target not found")
		}
		if len(s.State.HeroDeck) == 0 {
			return errors.New("deck empty")
		}
		s.State.DiscardPile = append(s.State.DiscardPile, target.CurrentHero)
		target.CurrentHero = s.State.HeroDeck[0]
		s.State.HeroDeck = s.State.HeroDeck[1:]
		target.CurrentAlibi = ""
		return nil
	})
}
