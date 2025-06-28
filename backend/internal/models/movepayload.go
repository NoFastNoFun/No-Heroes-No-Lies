package models

// MovePayload is what the client sends in POST /game/{id}/move.
// @Description Move payload for game actions
type MovePayload struct {
	Type          string   `json:"type" example:"demask" enums:"demask,fight,power"` // demask | fight | power
	TargetPlayer  string   `json:"target_player,omitempty" example:"player123"`      // demask
	GuessHero     string   `json:"guess,omitempty" example:"hero_name"`              // demask
	DeclaredAlibi string   `json:"declared_alibi,omitempty" example:"alibi_text"`    // fight | power
	MonsterID     string   `json:"monster_id,omitempty" example:"monster123"`        // fight
	PowerID       string   `json:"power_id,omitempty" example:"power456"`            // power
	DrawnCardID   string   `json:"drawn_card_id,omitempty" example:"card789"`        // power
	DiscardCardID string   `json:"discard_card_id,omitempty" example:"card101"`      // power
	Targets       []string `json:"targets,omitempty" example:"player1,player2"`      // generic
}
