ALTER TABLE game_hero_discard ADD COLUMN IF NOT EXISTS burned BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS challenger_player_id UUID REFERENCES players(id);
ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS challenger_strength_bonus INT NOT NULL DEFAULT 0;
ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS dual_attack_used BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS last_shoot_correct_player_id UUID REFERENCES players(id);
ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS last_shoot_correct BOOLEAN;
ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS mimicked_actor_id UUID REFERENCES players(id);
ALTER TABLE game_turn_state ADD COLUMN IF NOT EXISTS mimicked_hero_id TEXT REFERENCES heroes(id);
