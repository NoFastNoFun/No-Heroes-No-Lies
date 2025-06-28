package models

// MovePayload is what the client sends in POST /game/{id}/move.
type MovePayload struct {
	Type          string   `json:"type"`                      // demask | fight | power
	TargetPlayer  string   `json:"target_player,omitempty"`   // demask
	GuessHero     string   `json:"guess,omitempty"`           // demask
	DeclaredAlibi string   `json:"declared_alibi,omitempty"`  // fight | power
	MonsterID     string   `json:"monster_id,omitempty"`      // fight
	PowerID       string   `json:"power_id,omitempty"`        // power
	DrawnCardID   string   `json:"drawn_card_id,omitempty"`   // power
	DiscardCardID string   `json:"discard_card_id,omitempty"` // power
	Targets       []string `json:"targets,omitempty"`         // generic
}
