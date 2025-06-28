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

// Loot defines monster rewards.
type Loot struct {
	Coins int `json:"coins"`
	Gems  int `json:"gems"`
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
