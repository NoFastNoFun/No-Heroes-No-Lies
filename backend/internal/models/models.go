package models

// Power represents one active or passive ability.
// @Description Game power/ability definition
type Power struct {
	ID          string `json:"id" example:"power123"`
	Name        string `json:"name" example:"Fireball"`
	Type        string `json:"type" example:"active" enums:"active,passive"` // active | passive
	Cost        int    `json:"cost" example:"3"`
	Action      string `json:"action" example:"damage"`
	Target      string `json:"target" example:"enemy"`
	Trigger     string `json:"trigger,omitempty" example:"on_hit"`
	Order       int    `json:"order" example:"1"` // execution order within a card
	Description string `json:"description" example:"Deals 5 damage to target"`
}

// Loot defines monster rewards.
// @Description Loot rewards from defeating monsters
type Loot struct {
	Coins int `json:"coins" example:"10"`
	Gems  int `json:"gems" example:"2"`
}

// GameSession is the main session struct.
// @Description Game session information
type GameSession struct {
	ID        string    `json:"id" example:"session123"`
	PlayerIDs []string  `json:"player_ids" example:"player1,player2"`
	State     GameState `json:"state"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt string    `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string    `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}
