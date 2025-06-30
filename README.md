# No Heroes No Lies - Custom Game Server

Authoritative backend for the multiplayer bluff-and-battle card game **"No Heroes No Lies."**  
Pairs with **PocketBase** for Auth + DB + Realtime and deploys via **Coolify** on our VPS.

---

## Purpose

| Goal | Details |
|------|---------|
| **Fair play** | Validate every move, enforce costs, turn order, passives, lie-challenges, win rules |
| **Persistence** | Persist authoritative state in PocketBase (`game_sessions`, `moves`, ...) |
| **Simple API** | Lightweight HTTP endpoints for session creation, join, and authentication. All in-game actions use WebSocket. |
| **Zero-trust** | Clients only receive per-player "sanitized" views - hidden cards & decks stay server-side |

---

## Architecture

| Layer | Responsibility |
|-------|----------------|
| **PocketBase** | Collections (`users`, `game_sessions`, `moves`), authentication, realtime streams |
| **Go server**  | Game logic, lobby, power engine, AO_KEY header, host-gate, `/api/health`, **WebSocket game logic** |
| **React Frontend** | Modern web interface with PocketBase authentication and game sessions |

*Every server -> PB call includes **AO_KEY**; no admin token required.*

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

- `POST /api/game` - Create new game session
- `POST /api/game/{id}/join` - Join game session (with optional spectator parameter)
- `POST /api/auth/login` - Login
- `POST /api/auth/refresh` - Refresh session
- `GET /api/auth/logout` - Logout

---

## Feature Matrix

| Area | Status | Highlights |
|------|--------|------------|
| **Infrastructure** | OK | Host filter, AO_KEY header, JWT auth, health probe |
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
| **API polish** | `POST /game` (create lobby) & `/game/{id}/forfeit` finalise docs / errors |
| **Passive triggers (full)** | Add future triggers (`on_steal_attempt`, `monster_slain`, etc.) |
| **Realtime fan-out** | Push websocket / PB subscription hints when game ends or state updates |
| **Graceful shutdown** | Catch SIGTERM, drain connections |
| **Testing & CI** | Unit tests for power engine and challenge logic; GitHub/Coolify pipeline (`go vet`, `go test`, Docker build) |
| **Docs** | Public API schema (OpenAPI or MD) for client app |
| **Game Interface** | Implement the actual game UI when design is finalized |

---

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 16+
- PocketBase server running on `http://localhost:8090`

### Option 1: Development Scripts

**Windows:**

```cmd
start-dev.bat
```

**Linux/macOS:**

```bash
chmod +x start-dev.sh
./start-dev.sh
```

This will start both the backend server and frontend development server.

### Option 2: Manual Start

1. **Start PocketBase:**

```bash
# Download and run PocketBase
./pocketbase serve
```

2. **Start Backend:**

```bash
cd backend
go run ./cmd/server/main.go
```

3. **Start Frontend:**

```bash
cd frontend
# Create .env file with your configuration
echo "VITE_POCKETBASE_URL=http://localhost:8090" > .env
echo "VITE_API_URL=http://localhost:8080" >> .env
npm install
npm run dev
```

4. **Access the Application:**

- Backend API: <http://localhost:8080>
- PocketBase Admin: <http://localhost:8090/_/>
- API Documentation: <http://localhost:8080/swagger/>

### Option 3: Docker (Backend Only)

```bash
docker build -t nhnl .
docker run -p 8080:8080 \
           -e PORT=8080 \
           -e POCKETBASE_URL=http://localhost:8090 \
           -e POCKETBASE_AO_KEY=supersecret \
           -e ALLOWED_DOMAIN_SUFFIX=.example.com \
           nhnl
```

---

## Frontend Features

The new React frontend includes:

- **Real Authentication**: PocketBase-based user registration and login
- **Lobby System**: Browse and join available games from PocketBase
- **Game Sessions**: Create new games and manage sessions
- **Ready System**: Players can mark themselves as ready
- **Game Interface**: Placeholder for the actual game UI
- **Modern UI**: Built with React, TypeScript, and Tailwind CSS

### Frontend Tech Stack

- React 18 with TypeScript
- Vite for fast development
- React Router for navigation
- Tailwind CSS for styling
- Axios for API communication
- PocketBase SDK for authentication
- Lucide React for icons

### Authentication Flow

1. **Registration**: Users create accounts with email, username, and password
2. **Login**: Users authenticate with email and password
3. **Token Management**: PocketBase handles JWT tokens automatically
4. **Protected Routes**: All game routes require authentication
5. **Auto-login**: Auth state persists across page reloads

---

## API Usage

- **All in-game actions:** Use WebSocket `/ws/game/{id}`
- **REST:** Only for session creation, join, and authentication

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
