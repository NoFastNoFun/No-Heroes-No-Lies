# No Heroes No Lies - Custom Game Server

**Update 2024-07-03:**

- Fully migrated backend from PocketBase to PostgreSQL (persistent) and Redis (ephemeral/session)
- User registration and login with JWT auth (bcrypt, users table)
- All game/session logic, move logging, and state handled by new stack
- .env loading for local dev
- Game archival: finished games are saved in Postgres and become immutable
- All PocketBase code and references removed
- End-to-end flow: register/login → create/join game → play → finish → game is archived and cannot be modified

---

## Purpose

| Goal | Details |
|------|---------|
| **Fair play** | Validate every move, enforce costs, turn order, passives, lie-challenges, win rules |
| **Persistence** | Persist authoritative state in PostgreSQL (`users`, `cards`, `powers`, `moves`, `archived_games`) and Redis (sessions) |
| **Simple API** | Lightweight HTTP endpoints for session creation, join, and authentication. All in-game actions use WebSocket. |
| **Zero-trust** | Clients only receive per-player "sanitized" views - hidden cards & decks stay server-side |

---

## Architecture

| Layer | Responsibility |
|-------|----------------|
| **PostgreSQL** | Users, cards, powers, moves, archived_games |
| **Redis**      | Game sessions (ephemeral, fast access) |
| **Go server**  | Game logic, lobby, JWT auth, health probe, WebSocket game logic |
| **React Frontend** | Modern web interface with JWT authentication and game sessions |

---

## WebSocket Game Protocol

All in-game actions (moves, ready, forfeit, etc.) are handled via a single WebSocket connection:

- **Endpoint:** `/ws/game/{id}`
- **Authentication:** Uses the `game_auth` cookie (JWT)
- **Message format:**

```json
{
  "type": "move", // or "ready", "forfeit"
  "payload": { ... }
}
```

- **Supported types:**
  - `move`: Submit a move (demask, fight, power, etc.)
  - `ready`: Toggle ready status
  - `forfeit`: Forfeit the current game session

- **Server broadcasts:**
  - On any valid action, the updated game state is broadcast to all connected clients in the session.

---

## REST API (Lobby & Auth Only)

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/refresh` - Refresh session
- `GET /api/auth/logout` - Logout
- `POST /api/game` - Create new game session
- `POST /api/game/{id}/join` - Join game session (with optional spectator parameter)

---

## Feature Matrix

| Area | Status | Highlights |
|------|--------|------------|
| **Infrastructure** | OK | Host filter, JWT auth, health probe, .env loading |
| **Lobby system**   | OK | Join/leave before start, per-player **Ready/Not-ready**, start only when everyone ready, late joiners become **spectators** |
| **Session lifecycle** | OK | Create -> join -> start (2-15 players), deck build, burn, deal, random first player |
| **Moves & powers** | OK | `demask`, `fight`, full **26 active powers** + 3 passives, order-chain & cost enforcement, audit in `moves` |
| **Challenge window** | OK | 10-second lie challenge (configurable) with rollback |
| **Passives implemented** | OK | `alternate_strength`, `teamwork`, `keep_gems` (alibi-based) |
| **Combat & loot**  | OK | Strength compare, loot payout via teamwork, dual-attack flag |
| **Win detection**  | OK | Auto-end: last survivor **or** coins >= (players + 1); draw if everyone KO same turn |
| **Spectators**     | OK | `/join?spectator=1` after start; spectators cannot act |
| **Sanitized view** | OK | `/game/{id}` returns only info the caller is allowed to see |
| **Docker / Coolify** | OK | Multi-stage image, env vars, health-check endpoint |

---

## Remaining Work

| Group | To-do |
|-------|-------|
| **API polish** | Finalize docs / errors for all endpoints |
| **Game logic core** | Enforce power order chain, cost deduction, action execution, passive triggers, win conditions |
| **Testing & CI** | Unit/integration tests for full cycle, CI pipeline |
| **Docs** | Public API schema (OpenAPI or MD) for client app |
| **Frontend** | UI integration with new backend |

---

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 16+
- PostgreSQL and Redis running (see docker-compose.yml)

### Development

**Windows:**

```cmd
start-dev.bat
```

**Linux/macOS:**

```bash
chmod +x start-dev.sh
./start-dev.sh
```

### Manual Start

1. **Start PostgreSQL and Redis:**

```bash
docker compose up -d
```

2. **Start Backend:**

```bash
cd backend
go run ./cmd/server/main.go
```

3. **Start Frontend:**

```bash
cd frontend
npm install
npm run dev
```

4. **Access the Application:**

- Backend API: <http://localhost:8080>
- API Documentation: <http://localhost:8080/swagger/>

---

## API Usage

- **All in-game actions:** Use WebSocket `/ws/game/{id}`
- **REST:** Only for session creation, join, and authentication

---

Happy bluffing!

### Card Lifecycle

- **Discarded cards**: Hero cards that are played/discarded but not removed from the game. The last discarded card is public. Discarded cards can be recycled into the hero deck.
- **Burned cards**: Hero cards removed from the game entirely and hidden from memory. Burned cards are never recycled or made public.

### Power Move Enforcement

- When using a hero's power, a player **must** draw a card and discard either the drawn card or their current card. This is strictly enforced by the backend.

### Win Conditions

- The game ends when only one player remains alive, **or** when any player reaches a number of coins equal to the number of players plus one.

---

## Test Status

- All backend code builds and passes tests as of the latest update.
