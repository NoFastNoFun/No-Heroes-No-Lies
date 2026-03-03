package design

import "testing"

func TestGetCardHeroBasics(t *testing.T) {
	card, err := GetCard("troll")
	if err != nil {
		t.Fatalf("expected no error for known card, got %v", err)
	}
	if card.ID != "troll" {
		t.Fatalf("expected id troll, got %s", card.ID)
	}
	if card.Type != "hero" {
		t.Fatalf("expected type hero, got %s", card.Type)
	}
	if card.Name == "" {
		t.Fatal("expected non-empty name for hero card")
	}
	if card.Strength != 7 {
		t.Fatalf("expected strength 7 for troll, got %d", card.Strength)
	}
	if card.DefaultAmountPerSession != 2 {
		t.Fatalf("expected default amount per session 2 for troll, got %d", card.DefaultAmountPerSession)
	}
	if len(card.PowerIDs) != 2 {
		t.Fatalf("expected 2 powers for troll, got %d", len(card.PowerIDs))
	}
}

func TestGetCardMonsterBasics(t *testing.T) {
	card, err := GetCard("dragon")
	if err != nil {
		t.Fatalf("expected no error for known monster, got %v", err)
	}
	if card.Type != "monster" {
		t.Fatalf("expected type monster, got %s", card.Type)
	}
	if card.Strength != 9 {
		t.Fatalf("expected strength 9 for dragon, got %d", card.Strength)
	}
	if card.Loot.Coins != 1 || card.Loot.Gems != 5 {
		t.Fatalf("expected loot (1,5) for dragon, got (%d,%d)", card.Loot.Coins, card.Loot.Gems)
	}
}

func TestGetCardUnknown(t *testing.T) {
	card, err := GetCard("unknown_card_id")
	if err == nil {
		t.Fatalf("expected error for unknown card id, got nil (card=%+v)", card)
	}
}

func TestGetPowerBasics(t *testing.T) {
	pwr, err := GetPower("change_monster")
	if err != nil {
		t.Fatalf("expected no error for known power, got %v", err)
	}
	if pwr.ID != "change_monster" {
		t.Fatalf("expected id change_monster, got %s", pwr.ID)
	}
	if pwr.Name == "" {
		t.Fatal("expected non-empty name for power")
	}
	if pwr.Type != "active" {
		t.Fatalf("expected type active for change_monster, got %s", pwr.Type)
	}
	if pwr.Cost != 0 {
		t.Fatalf("expected cost 0 for change_monster, got %d", pwr.Cost)
	}
	if pwr.Target != "monster" {
		t.Fatalf("expected target monster for change_monster, got %s", pwr.Target)
	}
	if pwr.Action != "change_monster" {
		t.Fatalf("expected action change_monster, got %s", pwr.Action)
	}
}

func TestGetPowerPassive(t *testing.T) {
	pwr, err := GetPower("keep_gems")
	if err != nil {
		t.Fatalf("expected no error for known passive power, got %v", err)
	}
	if pwr.Type != "passive" {
		t.Fatalf("expected type passive for keep_gems, got %s", pwr.Type)
	}
	if pwr.Trigger != "steal_attempt" {
		t.Fatalf("expected trigger steal_attempt for keep_gems, got %s", pwr.Trigger)
	}
}

func TestGetPowerUnknown(t *testing.T) {
	pwr, err := GetPower("unknown_power")
	if err == nil {
		t.Fatalf("expected error for unknown power, got nil (power=%+v)", pwr)
	}
}

func TestAllCardsIncludesHeroesAndMonsters(t *testing.T) {
	all := AllCards()
	if len(all) < 30 {
		t.Fatalf("expected at least 30 cards (heroes + monsters), got %d", len(all))
	}
	var heroCount, monsterCount int
	for _, c := range all {
		switch c.Type {
		case "hero":
			heroCount++
		case "monster":
			monsterCount++
		default:
			t.Fatalf("unexpected card type %q for card %s", c.Type, c.ID)
		}
	}
	if heroCount != 15 {
		t.Fatalf("expected 15 heroes, got %d", heroCount)
	}
	if monsterCount != 18 {
		t.Fatalf("expected 18 monsters, got %d", monsterCount)
	}
}

