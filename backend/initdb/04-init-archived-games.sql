CREATE TABLE IF NOT EXISTS archived_games (
    id VARCHAR PRIMARY KEY,
    player_ids TEXT[] NOT NULL,
    state JSONB NOT NULL,
    moves JSONB NOT NULL,
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);