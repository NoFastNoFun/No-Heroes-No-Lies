## Backend Operation Modes

### Full mode

In full mode, the backend runs with:

- Redis available for live session and game state.
- Postgres available for user accounts, move logs, and archived games.
- Static design data embedded in the `internal/design` package for all cards and powers.

Behavior:

- User registration, login, and refresh endpoints are enabled.
- Moves are logged to Postgres and games can be archived and queried later.
- Core gameplay (sessions, moves, WebSockets) uses Redis and design data.

### DB-lite mode

In DB-lite mode, the backend runs with:

- Redis available for live session and game state.
- Postgres unavailable or misconfigured.
- Static design data from `internal/design` for all cards and powers.

Behavior:

- Core gameplay works normally: sessions can be created, joined, started, and played.
- User registration and login endpoints return HTTP 503 with a clear error message.
- Move logging and archived game history are effectively disabled (no Postgres).

### Configuration

- **Redis**
  - Configured via the existing Redis address in the backend configuration.
  - Must point to a reachable Redis instance; otherwise the server will exit.

- **Postgres**
  - Initialized via environment variables consumed by `internal/db`:
    - `POSTGRES_HOST`
    - `POSTGRES_PORT`
    - `POSTGRES_USER`
    - `POSTGRES_PASSWORD`
    - `POSTGRES_DB`
    - `POSTGRES_SSLMODE` (optional, defaults to `disable`)
  - If initialization fails at startup, the server logs a warning and continues in DB-lite mode.

### Observability

- On startup, the server logs which mode is active:
  - Full mode if Postgres initialization succeeds.
  - DB-lite mode with a warning if Postgres initialization fails.

