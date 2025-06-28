# No Heroes No Lies - Custom Game Server

Authoritative backend for the multiplayer bluff-and-battle card game **"No Heroes No Lies."**  
Pairs with **PocketBase** for Auth + DB + Realtime and deploys via **Coolify** on our VPS.

---

## Purpose

| Goal | Details |
|------|---------|
| **Fair play** | Validate every move, enforce costs, turn order, passives, lie-challenges, win rules |
| **Persistence** | Persist authoritative state in PocketBase (`game_sessions`, `moves`, ...) |
| **Simple API** | Lightweight HTTP endpoints (`/game`, `/move`, ...) for any web / mobile client |
| **Zero-trust** | Clients only receive per-player "sanitized" views - hidden cards & decks stay server-side |

---

## Architecture

| Layer | Responsibility |
|-------|----------------|
| **PocketBase** | Collections (`games_accounts`, `cards`, `powers`, `game_sessions`, `moves`), realtime streams |
| **Go server**  | Game logic, lobby, power engine, AO_KEY header, host-gate, `/api/health` |
| **Front-end**  | REST calls to server, realtime updates from PocketBase |

*Every server -> PB call includes **AO_KEY**; no admin token required.*

---

## Feature Matrix

| Area | Status | Highlights |
|------|--------|------------|
| **Infrastructure** | OK | Host filter, AO_KEY header, JWT auth, health probe |
| **Lobby system**   | OK | Join/leave before start, per-player **Ready/Not-ready**, start only when everyone ready, late joiners become **spectators** |
| **Session lifecycle** | OK | Create -> join -> start (2-15 players), deck build, burn, deal, random first player |
| **Moves & powers** | OK | `demask`, `fight`, full **26 active powers** + 3 passives, order-chain & cost enforcement, audit in `moves` |
| **Challenge window** | OK | 5-second lie challenge with rollback |
| **Passives implemented** | OK | `alternate_strength`, `teamwork`, `keep_gems` (alibi-based) |
| **Combat & loot**  | OK | Strength compare, loot payout via teamwork, dual-attack flag |
| **Win detection**  | OK | Auto-end: last survivor **or** >= 5 coins; draw if everyone KO same turn |
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

---

## Running Locally

```bash
docker build -t nhnl .
docker run -p 8080:8080 \
           -e PORT=8080 \
           -e POCKETBASE_URL=http://localhost:8090 \
           -e POCKETBASE_AO_KEY=supersecret \
           -e ALLOWED_DOMAIN_SUFFIX=.example.com \
           nhnl
```

Clients authenticate with PocketBase JWT:

```http
POST /game/{id}/move
Authorization: Bearer <PB-JWT>
Content-Type: application/json
```

and subscribe to `game_sessions/{id}` via PocketBase realtime for live updates.

Happy bluffing
