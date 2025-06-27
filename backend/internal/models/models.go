package models

// Power represents one active or passive ability.
type Power struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // active | passive
	Cost        int    `json:"cost"`
	Action      string `json:"action"`
	Target      string `json:"target"`
	Trigger     string `json:"trigger,omitempty"`
	Order       int    `json:"order"` // execution order within a card
	Description string `json:"description"`
}

// Card is a hero, monster, etc.
type Card struct {
	ID                      string   `json:"id"`
	Type                    string   `json:"type"` // hero | monster
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	Strength                int      `json:"strength"`
	PowerIDs                []string `json:"power_ids"`
	Loot                    Loot     `json:"loot"`
	DefaultAmountPerSession int      `json:"default_amount_per_session"`
	CoverURL                string   `json:"cover"` // PB returns file URL
}

// Loot defines monster rewards.
type Loot struct {
	Coins int `json:"coins"`
	Gems  int `json:"gems"`
}

// PlayerState lives inside game session state.
type PlayerState struct {
	ID           string   `json:"id"`
	Life         int      `json:"life"`
	Coins        int      `json:"coins"`
	Gems         int      `json:"gems"`
	Glory        int      `json:"glory"`
	CurrentHero  string   `json:"current_hero"`
	CurrentAlibi string   `json:"current_alibi"`
	Hand         []string `json:"hand,omitempty"`
}

// GameState is the authoritative JSON blob.
type GameState struct {
	Players     []PlayerState `json:"players"`
	TurnOrder   []string      `json:"turn_order"`
	CurrentTurn string        `json:"current_turn"`
	CoinPool    int           `json:"coin_pool"`
	GemPool     int           `json:"gem_pool"`
}

// GameSession mirrors the PocketBase record we work with.
type GameSession struct {
	ID        string    `json:"id"`
	PlayerIDs []string  `json:"player_ids"`
	State     GameState `json:"state"`
	IsActive  bool      `json:"is_active"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
