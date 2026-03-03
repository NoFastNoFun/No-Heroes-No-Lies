package services

import (
	"testing"

	"no-heroes-no-lies/internal/models"
)

func TestApplyPowerDrawAndDiscard_DiscardCurrentHero(t *testing.T) {
	// Set up a minimal game session with a hero deck and one player.
	session := &models.GameSession{
		State: models.GameState{
			HeroDeck: []string{"troll", "elf"},
		},
	}
	player := &models.PlayerState{
		ID:          "p1",
		CurrentHero: "mage",
	}
	session.State.Players = []models.PlayerState{*player}

	svc := &GameService{}

	move := &models.MovePayload{
		// Player chooses to discard their current hero, so new hero becomes drawn card.
		DiscardCardID: "mage",
	}

	if err := svc.applyPowerDrawAndDiscard(session, player, move); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := move.DrawnCardID; got != "troll" {
		t.Fatalf("expected drawn_card_id troll, got %s", got)
	}
	if got := player.CurrentHero; got != "troll" {
		t.Fatalf("expected current hero to switch to drawn card troll, got %s", got)
	}
	if len(session.State.HeroDeck) != 1 || session.State.HeroDeck[0] != "elf" {
		t.Fatalf("expected hero deck to now contain [elf], got %#v", session.State.HeroDeck)
	}
	if got := session.State.PublicDiscard; got != "mage" {
		t.Fatalf("expected public discard to be mage, got %s", got)
	}
	if len(session.State.DiscardPile) != 1 || session.State.DiscardPile[0] != "mage" {
		t.Fatalf("expected discard pile to contain [mage], got %#v", session.State.DiscardPile)
	}
}

func TestApplyPowerDrawAndDiscard_DiscardDrawnCard(t *testing.T) {
	// Set up a minimal game session with a hero deck and one player.
	session := &models.GameSession{
		State: models.GameState{
			HeroDeck: []string{"troll", "elf"},
		},
	}
	player := &models.PlayerState{
		ID:          "p1",
		CurrentHero: "mage",
	}
	session.State.Players = []models.PlayerState{*player}

	svc := &GameService{}

	move := &models.MovePayload{
		// Player chooses to discard the drawn card, keeping their current hero.
		DiscardCardID: "troll",
	}

	if err := svc.applyPowerDrawAndDiscard(session, player, move); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := move.DrawnCardID; got != "troll" {
		t.Fatalf("expected drawn_card_id troll, got %s", got)
	}
	if got := player.CurrentHero; got != "mage" {
		t.Fatalf("expected current hero to remain mage, got %s", got)
	}
	if len(session.State.HeroDeck) != 1 || session.State.HeroDeck[0] != "elf" {
		t.Fatalf("expected hero deck to now contain [elf], got %#v", session.State.HeroDeck)
	}
	if got := session.State.PublicDiscard; got != "troll" {
		t.Fatalf("expected public discard to be troll, got %s", got)
	}
	if len(session.State.DiscardPile) != 1 || session.State.DiscardPile[0] != "troll" {
		t.Fatalf("expected discard pile to contain [troll], got %#v", session.State.DiscardPile)
	}
}

