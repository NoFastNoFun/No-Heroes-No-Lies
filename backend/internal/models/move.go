package models

// Move mirrors the `moves` collection.
type Move struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	PlayerID  string `json:"player_id"`
	Type      string `json:"type"`      // demask | fight | power
	Data      string `json:"move_data"` // raw JSON payload
	CreatedAt string `json:"created_at"`
}
