package models

// Card mirrors the cards collection.
type Card struct {
	ID                      string   `json:"id"`
	Type                    string   `json:"type"`
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	Strength                int      `json:"strength"`
	PowerIDs                []string `json:"power_ids"`
	Loot                    Loot     `json:"loot"`
	DefaultAmountPerSession int      `json:"default_amount_per_session"`
	CoverURL                string   `json:"cover"`
}
