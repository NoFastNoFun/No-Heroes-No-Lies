package view

import "no-heroes-no-lies/internal/models"

// SanitizedSelf is what the requesting player sees about themselves.
// @Description Player's own information in game view
type SanitizedSelf struct {
	ID           string `json:"id" example:"player123"`
	CurrentHero  string `json:"current_hero" example:"hero_name"`
	CurrentAlibi string `json:"current_alibi" example:"alibi_text"`
	Life         int    `json:"life" example:"10"`
	Coins        int    `json:"coins" example:"5"`
	Gems         int    `json:"gems" example:"2"`
}

// SanitizedPlayer is what they see about every other player.
// @Description Other player information in game view
type SanitizedPlayer struct {
	ID    string `json:"id" example:"player456"`
	Life  int    `json:"life" example:"8"`
	Coins int    `json:"coins" example:"3"`
	Gems  int    `json:"gems" example:"1"`
}

// SessionView is the JSON sent to the client.
// @Description Complete game session view for a player
type SessionView struct {
	SessionID      string            `json:"session_id" example:"session123"`
	You            SanitizedSelf     `json:"you"`
	Others         []SanitizedPlayer `json:"others"`
	ActiveMonsters []string          `json:"active_monsters" example:"monster1,monster2"`
	CurrentTurn    string            `json:"current_turn" example:"player123"`
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
