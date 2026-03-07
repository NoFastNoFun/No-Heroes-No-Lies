ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS picked_hero_id TEXT REFERENCES heroes(id);
