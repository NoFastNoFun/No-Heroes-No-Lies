# No Heroes No Lies - Backend

Go backend for the NHML card game. Temporary (username-only) sessions, REST lobby, WebSocket in-game actions.

## Run with Docker

From repo root:

```bash
docker-compose up --build
```

API: http://localhost:8080

## Run locally

Requires PostgreSQL (e.g. port 5432, database `nhml`, user/pass `postgres/postgres`).

```bash
cd backend
go build -o server ./cmd/server
./server
```

Or set `DATABASE_URL` and run `go run ./cmd/server`.

## Env

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | HTTP port |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/nhml?sslmode=disable | Postgres connection |
| JWT_SECRET | dev-secret-change-in-production | Session token signing |
| SESSION_TTL | 24h | Session expiry |
| CHALLENGE_WINDOW | 10s | Accusation window after declare |

## API

- `POST /api/v1/session` — body `{"username":"..."}` → session token (no account).
- `POST /api/v1/games` — create game (auth).
- `GET /api/v1/games` — list lobby games (auth).
- `GET /api/v1/games/{id}` — game detail (auth).
- `GET /api/v1/games/{id}/state` — game state for player (auth).
- `POST /api/v1/games/{id}/join` — join game (auth).
- `POST /api/v1/games/{id}/start` — start game, master only (auth).
- `GET /api/v1/games/{id}/ws` — WebSocket, query `?token=...` or `Authorization: Bearer <token>`. Send JSON `{"type":"pick"}`, `{"type":"discard","payload":{"keep_picked":true}}`, `{"type":"declare","payload":{"hero_id":"mage"}}`, `{"type":"power","payload":{...}}`, `{"type":"attack","payload":{"monster_slot":0}}`, `{"type":"demask","payload":{...}}`, `{"type":"accuse","payload":{...}}`.

## Health

- `GET /health` — liveness.
- `GET /ready` — readiness (DB ping).
