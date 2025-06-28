# No Heroes No Lies - Custom Game Server

Authoritative backend that enforces **all** rules for the multiplayer bluff-and-battle card game **"No Heroes No Lies."**  
Runs beside **PocketBase** (Auth + DB + Realtime) and is deployed with **Coolify** on our VPS.

---

## 🎯 Purpose

| Goals | Details |
|-------|---------|
| **Fair play** | Validate every move, enforce costs, turn order, passives, challenges, win conditions |
| **Persistence** | Store authoritative state in PocketBase (`game_sessions`, `moves`, …) |
| **Simple API** | Lightweight HTTP endpoints (`/game`, `/move`, …) consumable by any web/mobile client |
| **Zero trust** | Clients never see hidden cards or decks; server derives per-player "sanitized" views |

---

## ⚙️ Architecture

| Layer | Responsibility |
|-------|----------------|
| **PocketBase** | Accounts (`games_accounts`), collections (`cards`, `powers`, `game_sessions`, `moves`), realtime feeds |
| **Go server** | All game logic, power engine, host-gate, AO_KEY header, health probe |
| **Front-end** | Calls the API, listens to PB realtime, renders UI |

*Server → PocketBase requests always include `AO_KEY`; no admin token needed.*

---

## ✅ Current Feature Matrix

| Area | Status | Highlights |
|------|--------|------------|
| **Infrastructure** | ✅ | Host filter, AO_KEY header, JWT auth middleware, `/api/health` |
| **Session lifecycle** | ✅ | Create, join, start, deck build, burn, initial deal, random first player |
| **Moves & logging** | ✅ | `demask`, `fight`, full **26-power** registry, move audit in `moves` |
| **Challenge window** | ✅ | 5-second lie challenge with full rollback |
| **Passives** | 🟡 | `alternate_strength`, `teamwork`, `keep_gems` implemented; hook helpers in place |
| **Combat & loot** | ✅ | Strength compare, loot payout, deck refill, dual-attack support |
| **Cost & order chains** | ✅ | Gem deduction and order-0/1 enforcement |
| **Anti-cheat view** | ✅ | `/game/{id}` returns player-specific sanitized JSON |
| **Docker / Coolify** | ✅ | Static binary, AO_KEY/Host env vars, health-check |

---

## 🗂️ What's Left

| Task Group | Remaining Work |
|------------|----------------|
| **API polish** | `POST /game` & `POST /game/{id}/forfeit` (public endpoints already scaffolded) |
| **Passive triggers (full)** | • Hook "teamwork" into **coin** gains • Add event hook map for future triggers (`on_steal_attempt`, `monster_slain`, etc.) |
| **Mimic & complex powers** | Finish `mimic_power` (copy target's order-0 active) and any TODO stubs |
| **Win / lose conditions** | • First to 4 glory **or** last player with life > 0 ends game • Persist winner in session |
| **Realtime fan-out** | Use PB realtime or lightweight WS push after each state update |
| **Graceful shutdown** | Capture SIGTERM, close HTTP server cleanly |
| **Testing** | Unit tests for power engine & challenge logic; integration tests with PB dev instance |
| **CI / lint** | GitHub or Coolify pipeline: `go vet`, `go test`, Docker build |
| **Docs** | API schema (OpenAPI or markdown) for client developers |

---

## 🛣️ Next Suggested Milestone

> **Win-Detection & Game-End:**  
> • track glory points on monster slay / duel victory  
> • check win condition after every state change  
> • mark `is_active=false`, declare winner(s), push realtime event.

Once that's in place the game loop is complete; later tasks are quality-of-life (realtime polish, tests, CI).

---

## 📌 Running Locally

```bash
# build and run
docker build -t nhnl .
docker run -e PORT=8080 \
           -e POCKETBASE_URL=http://localhost:8090 \
           -e POCKETBASE_AO_KEY=YOUR_SECRET \
           -e ALLOWED_DOMAIN_SUFFIX=.example.com \
           -p 8080:8080 nhnl
````

---

Clients authenticate with their PocketBase JWT:

```http
POST /game/{id}/move
Authorization: Bearer <PB-JWT>
Content-Type: application/json
```

and listen to PocketBase realtime channel `game_sessions/{id}` for live updates.

Happy bluffing! 🎲
