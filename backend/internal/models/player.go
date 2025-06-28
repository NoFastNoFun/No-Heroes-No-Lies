package models

type PlayerState struct {
	ID            string   `json:"id"`
	Life          int      `json:"life"`
	Coins         int      `json:"coins"`
	Gems          int      `json:"gems"`
	Glory         int      `json:"glory"`
	CurrentHero   string   `json:"current_hero"`
	CurrentAlibi  string   `json:"current_alibi"`
	Hand          []string `json:"hand,omitempty"`
	BonusStrength int      `json:"bonus_strength"` // temp +1 from add_strength_to_challenger
	AltStrength   bool     `json:"alt_strength"`   // For werewolf
}
