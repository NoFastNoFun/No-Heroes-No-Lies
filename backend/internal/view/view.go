package view

import "no-heroes-no-lies/internal/models"

// SanitizedSelf is what the requesting player sees about themselves.
type SanitizedSelf struct {
	ID           string `json:"id"`
	CurrentHero  string `json:"current_hero"`
	CurrentAlibi string `json:"current_alibi"`
	Life         int    `json:"life"`
	Coins        int    `json:"coins"`
	Gems         int    `json:"gems"`
}

// SanitizedPlayer is what they see about every other player.
type SanitizedPlayer struct {
	ID    string `json:"id"`
	Life  int    `json:"life"`
	Coins int    `json:"coins"`
	Gems  int    `json:"gems"`
}

// SessionView is the JSON sent to the client.
type SessionView struct {
	SessionID      string            `json:"session_id"`
	You            SanitizedSelf     `json:"you"`
	Others         []SanitizedPlayer `json:"others"`
	ActiveMonsters []string          `json:"active_monsters"`
	CurrentTurn    string            `json:"current_turn"`
}

// Build returns a per-player view of the session state.
func Build(session models.GameSession, playerID string) SessionView {
	var you SanitizedSelf
	var others []SanitizedPlayer

	for _, p := range session.State.Players {
		if p.ID == playerID {
			you = SanitizedSelf{
				ID:           p.ID,
				CurrentHero:  p.CurrentHero,
				CurrentAlibi: p.CurrentAlibi,
				Life:         p.Life,
				Coins:        p.Coins,
				Gems:         p.Gems,
			}
		} else {
			others = append(others, SanitizedPlayer{
				ID:    p.ID,
				Life:  p.Life,
				Coins: p.Coins,
				Gems:  p.Gems,
			})
		}
	}

	return SessionView{
		SessionID:      session.ID,
		You:            you,
		Others:         others,
		ActiveMonsters: session.State.ActiveMonsters,
		CurrentTurn:    session.State.CurrentTurn,
	}
}
