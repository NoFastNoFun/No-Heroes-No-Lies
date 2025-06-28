package models

// LastMove is stored for 5 s, so others can challenge.
type LastMove struct {
	ActorID       string `json:"actor_id"`
	Type          string `json:"type"` // fight | power
	DeclaredAlibi string `json:"declared_alibi"`
	MonsterID     string `json:"monster_id"`  // fight
	CoinsDelta    int    `json:"coins_delta"` // +loot or 0
	GemsDelta     int    `json:"gems_delta"`
	LifeDeltaID   string `json:"life_delta_id"` // whose life changed
	LifeDelta     int    `json:"life_delta"`    // −1 or +0
	ExpiresAt     int64  `json:"expires_at"`    // unix ms
}

// GameState …
type GameState struct {
	Players        []PlayerState   `json:"players"`
	TurnOrder      []string        `json:"turn_order"`
	CurrentTurn    string          `json:"current_turn"`
	CoinPool       int             `json:"coin_pool"`
	GemPool        int             `json:"gem_pool"`
	HeroDeck       []string        `json:"hero_deck"`
	MonsterDeck    []string        `json:"monster_deck"`
	ActiveMonsters []string        `json:"active_monsters"`
	Seed           int64           `json:"seed"`
	LastMove       *LastMove       `json:"last_move,omitempty"`
	DiscardPile    []string        `json:"discard_pile,omitempty"`   // discarded hero cards (non-burned)
	Order0UsedBy   map[string]bool `json:"order0_used_by,omitempty"` // tracks who used order-0 power this turn
	DualAttack     bool            `json:"dual_attack"`              // flag set by dual_attack power
}
