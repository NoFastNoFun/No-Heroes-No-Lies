CREATE TABLE heroes (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	strength INT NOT NULL,
	default_spawn_amount INT NOT NULL,
	power1 TEXT NOT NULL,
	power2 TEXT NOT NULL,
	rarity TEXT NOT NULL
);

CREATE TABLE monsters (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	strength INT NOT NULL,
	loot_coins INT NOT NULL,
	loot_gems INT NOT NULL,
	rarity TEXT NOT NULL,
	spawn_weight INT NOT NULL
);

CREATE TABLE powers (
	name TEXT PRIMARY KEY,
	cost_gems INT NOT NULL,
	target_type TEXT NOT NULL,
	is_passive BOOLEAN NOT NULL DEFAULT false,
	trigger_event TEXT,
	description TEXT
);
