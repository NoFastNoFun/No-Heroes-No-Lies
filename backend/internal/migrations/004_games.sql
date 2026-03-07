CREATE TABLE games (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	creator_id UUID NOT NULL REFERENCES players(id),
	status TEXT NOT NULL DEFAULT 'lobby' CHECK (status IN ('lobby', 'playing', 'finished')),
	max_players INT NOT NULL DEFAULT 4,
	current_turn_index INT NOT NULL DEFAULT 0,
	round_number INT NOT NULL DEFAULT 0,
	win_condition_met BOOLEAN NOT NULL DEFAULT false,
	last_discard_hero_id TEXT REFERENCES heroes(id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE game_players (
	game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
	player_id UUID NOT NULL REFERENCES players(id),
	slot_index INT NOT NULL,
	is_game_master BOOLEAN NOT NULL DEFAULT false,
	life INT NOT NULL DEFAULT 3,
	coins INT NOT NULL DEFAULT 0,
	gems INT NOT NULL DEFAULT 0,
	hero_id TEXT REFERENCES heroes(id),
	declared_hero_id TEXT REFERENCES heroes(id),
	is_wolf_form BOOLEAN NOT NULL DEFAULT false,
	is_eliminated BOOLEAN NOT NULL DEFAULT false,
	joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (game_id, player_id),
	UNIQUE (game_id, slot_index)
);

CREATE TABLE game_turn_state (
	game_id UUID PRIMARY KEY REFERENCES games(id) ON DELETE CASCADE,
	phase TEXT NOT NULL CHECK (phase IN ('pick', 'discard', 'declare', 'power', 'attack', 'demask', 'resolving_passives', 'between_turns')),
	phase_entered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	challenge_window_ends_at TIMESTAMPTZ,
	pending_action_type TEXT,
	pending_action_payload JSONB,
	acting_player_id UUID REFERENCES players(id),
	power_step TEXT CHECK (power_step IN ('power1_pending', 'power1_done', 'power2_pending'))
);

CREATE TABLE game_hero_deck (
	game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
	hero_id TEXT NOT NULL REFERENCES heroes(id),
	position INT NOT NULL,
	burned BOOLEAN NOT NULL DEFAULT false,
	PRIMARY KEY (game_id, position)
);

CREATE TABLE game_hero_discard (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
	hero_id TEXT NOT NULL REFERENCES heroes(id),
	discarded_by_player_id UUID REFERENCES players(id),
	discarded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_game_hero_discard_game ON game_hero_discard(game_id);

CREATE TABLE game_monster_deck (
	game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
	monster_id TEXT NOT NULL REFERENCES monsters(id),
	position INT NOT NULL,
	PRIMARY KEY (game_id, position)
);

CREATE TABLE game_active_monsters (
	game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
	monster_id TEXT NOT NULL REFERENCES monsters(id),
	slot_index INT NOT NULL CHECK (slot_index IN (0, 1)),
	PRIMARY KEY (game_id, slot_index)
);

CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_game_players_game ON game_players(game_id);
