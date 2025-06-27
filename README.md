# No Heroes No Lies - Custom Game Server

Authoritative backend that enforces all rules for the multiplayer card game **"No Heroes No Lies."**  
Runs alongside **PocketBase** (DB + Auth + Realtime)

---

## 🎯 Purpose

- Validate and process player moves securely  
- Enforce turns, costs, hidden-role bluffing, passives, win conditions  
- Persist game state in PocketBase  
- Offer a clean HTTP API for clients (web / mobile)  

---

## ⚙️ Architecture

| Layer            | Responsibility |
|------------------|----------------|
| **PocketBase**   | Auth (`games_accounts`), storage (`powers`, `cards`, `game_sessions`, `moves`), realtime subscriptions |
| **Go server**    | All game logic, HTTP API, host-gate, AO_KEY header, health probe |

**Server-PocketBase trust:** every request carries `AO_KEY` in the header; no admin API token needed.

---

## ✅ Feature Checklist / TODO

### 1. Infrastructure & Security

| Task | Status |
|------|--------|
| Host-filter middleware (`Host` must end with configured suffix) | **DONE** |
| Global AO_KEY header on all PB calls | **DONE** |
| Auth middleware (verify PocketBase user JWT, inject playerID) | **DONE** |
| `/api/health` liveness endpoint | **DONE** |

---

### 2. Data Models (PocketBase)

| Collection | Status | Notes |
|------------|--------|-------|
| `games_accounts` | **DONE** | custom auth table with status & avatar |
| `powers` | **DONE** | includes `order`, `type`, `cost`, `trigger` |
| `cards` | **DONE** | references `power_ids`, has `default_amount_per_session` |
| `game_sessions` | **DONE** | JSON state template implemented |
| `moves` | OPTIONAL | audit / replay - **TODO** to write from server |

---

### 3. API Endpoints

| Route | Status | Notes |
|-------|--------|-------|
| `GET  /api/health` | **DONE** | liveness probe |
| `GET  /game/{id}` | **DONE** | fetch current session |
| `POST /game/{id}/move` | **DONE (minimal)** | applies a power (rule-set still WIP) |
| `POST /game` | **TODO** | create session, build deck respecting card limits |
| `POST /game/{id}/forfeit` | **TODO** | player quits / surrender |

---

### 4. Game Logic Core

| Task | Status |
|------|--------|
| Turn-order enforcement | **DONE (basic)** |
| Cost deduction (coins / gems) | **TODO** |
| Power **order** chain enforcement | **TODO** |
| Passive trigger evaluation | **TODO** |
| Combat & loot (`strength`, `loot`) | **TODO** |
| Lying / hero-reveal mechanics | **TODO** |
| Win detection (glory ≥ 4 or last standing) | **TODO** |
| Persist each move to `moves` collection | **TODO** |

---

### 5. Real-Time Updates

| Task | Status |
|------|--------|
| Push state changes via PocketBase subscriptions | **TODO** |
| Optionally WebSocket fan-out for low-latency UX | **TODO** |

---

### 6. Deployment / Ops

| Task | Status |
|------|--------|
| Multi-stage Dockerfile (static binary, Alpine) | **DONE** |
| Coolify service with env vars: `PORT`, `POCKETBASE_URL`, `POCKETBASE_AO_KEY`, `ALLOWED_DOMAIN_SUFFIX` | **DONE** |
| `/health` used as Docker/Coolify health-check | **DONE** |
| Graceful shutdown | **TODO** (signal handling) |

---

### 7. Testing

| Area | Status |
|------|--------|
| Unit tests for `services/game.go` (turns, costs, passives) | **TODO** |
| Integration tests against PocketBase sandbox | **TODO** |

---

## 💡 Notes

- All non-game concerns (auth, host gate, health, PB headers) are finished.  
- Remaining work is **pure gameplay logic** plus optional niceties (realtime fan-out, tests).  
- Clients send `Authorization: Bearer <PB-JWT>` and interact only with `/game/*` endpoints; they listen to PocketBase realtime feeds for state refresh.
