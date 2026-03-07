ALTER TABLE games ADD COLUMN winner_id UUID REFERENCES players(id);
