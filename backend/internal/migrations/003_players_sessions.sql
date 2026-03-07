CREATE TABLE players (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	token_hash TEXT NOT NULL UNIQUE,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
