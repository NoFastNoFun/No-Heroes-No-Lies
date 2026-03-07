package repository

import (
	"context"
	"math/rand"
)

type HeroRow struct {
	ID                 string
	Name               string
	Strength           int
	DefaultSpawnAmount int
	Power1             string
	Power2             string
	Rarity             string
}

func (db *DB) GetAllHeroes(ctx context.Context) ([]HeroRow, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, name, strength, default_spawn_amount, power1, power2, rarity FROM heroes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HeroRow
	for rows.Next() {
		var h HeroRow
		if err := rows.Scan(&h.ID, &h.Name, &h.Strength, &h.DefaultSpawnAmount, &h.Power1, &h.Power2, &h.Rarity); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (db *DB) GetHeroByID(ctx context.Context, heroID string) (*HeroRow, error) {
	var h HeroRow
	err := db.Pool.QueryRow(ctx,
		`SELECT id, name, strength, default_spawn_amount, power1, power2, rarity FROM heroes WHERE id = $1`,
		heroID,
	).Scan(&h.ID, &h.Name, &h.Strength, &h.DefaultSpawnAmount, &h.Power1, &h.Power2, &h.Rarity)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

type MonsterRow struct {
	ID         string
	Name       string
	Strength   int
	LootCoins  int
	LootGems   int
	Rarity     string
	SpawnWeight int
}

func (db *DB) GetAllMonsters(ctx context.Context) ([]MonsterRow, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, name, strength, loot_coins, loot_gems, rarity, spawn_weight FROM monsters ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MonsterRow
	for rows.Next() {
		var m MonsterRow
		if err := rows.Scan(&m.ID, &m.Name, &m.Strength, &m.LootCoins, &m.LootGems, &m.Rarity, &m.SpawnWeight); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (db *DB) GetMonsterByID(ctx context.Context, monsterID string) (*MonsterRow, error) {
	var m MonsterRow
	err := db.Pool.QueryRow(ctx,
		`SELECT id, name, strength, loot_coins, loot_gems, rarity, spawn_weight FROM monsters WHERE id = $1`,
		monsterID,
	).Scan(&m.ID, &m.Name, &m.Strength, &m.LootCoins, &m.LootGems, &m.Rarity, &m.SpawnWeight)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// DrawRandomMonsterByWeight returns one monster ID according to spawn_weight (infinite deck by percentage).
func (db *DB) DrawRandomMonsterByWeight(ctx context.Context) (string, error) {
	monsters, err := db.GetAllMonsters(ctx)
	if err != nil || len(monsters) == 0 {
		return "", err
	}
	var total int
	for _, m := range monsters {
		total += m.SpawnWeight
	}
	if total <= 0 {
		return monsters[0].ID, nil
	}
	r := rand.Intn(total)
	for _, m := range monsters {
		r -= m.SpawnWeight
		if r < 0 {
			return m.ID, nil
		}
	}
	return monsters[len(monsters)-1].ID, nil
}
