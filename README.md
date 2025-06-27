# No Heroes No Lies – Custom Game Server

Authoritative backend that enforces all rules for the multiplayer card game **“No Heroes No Lies.”**  
Runs alongside **PocketBase** (DB + Auth + Realtime) and is deployed with **Coolify** on a VPS.

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
| **Go server**    | Game logic, HTTP API, host-gate, AO_KEY header, health probe |

*All PocketBase calls carry the mandatory `AO_KEY` header; no admin token is used.*

---

## ✅ Feature Checklist / TODO

### 1. Infrastructure & Security

| Task | Status |
|------|--------|
| Host-filter middleware (`Host` must end with configured suffix) | ✅ |
| Global `AO_KEY` header on all PB calls | ✅ |
| Auth middleware (verify PocketBase user JWT, inject playerID) | ✅ |
| `/api/health` liveness endpoint | ✅ |

---

### 2. Data Models (PocketBase)

| Collection | Status | Notes |
|------------|--------|-------|
| `games_accounts` | ✅ | custom auth table with status & avatar |
| `powers` | ✅ | has `order`, `type`, `cost`, `trigger` |
| `cards` | ✅ | references `power_ids`, includes `default_amount_per_session` |
| `game_sessions` | ✅ | JSON state template implemented |
| `moves` | ✅ | action log for audit / replay |

---

### 3. API Endpoints

| Route | Status | Notes |
|-------|--------|-------|
| `GET  /api/health` | ✅ | liveness probe |
| `GET  /game/{id}` | ✅ | fetch current session |
| `POST /game/{id}/move` | **✅ (minimal)** | applies a power (rule-set WIP) |
| `POST /game` | **TODO** | create session, build deck with card limits |
| `POST /game/{id}/forfeit` | **TODO** | player quits / surrender |

---

### 4. Game Logic Core

| Task | Status |
|------|--------|
| Turn-order enforcement | **✅ (basic)** |
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
| Optional WebSocket fan-out for low-latency UX | **TODO** |

---

### 6. Deployment / Ops

| Task | Status |
|------|--------|
| Multi-stage Dockerfile (static binary, Alpine) | ✅ |
| Coolify service with env vars (`PORT`, `POCKETBASE_URL`, `POCKETBASE_AO_KEY`, `ALLOWED_DOMAIN_SUFFIX`) | ✅ |
| `/api/health` used as Docker/Coolify health-check | ✅ |
| Graceful shutdown | **TODO** (signal handling) |

---

### 7. Testing

| Area | Status |
|------|--------|
| Unit tests for `services/game.go` (turns, costs, passives) | **TODO** |
| Integration tests against PocketBase sandbox | **TODO** |

---

## 💡 Notes

- All infrastructure (auth, host gate, health, PocketBase wiring) is in place.  

- Remaining work is purely gameplay logic plus optional realtime fan-out and tests.  
- Clients send `Authorization: Bearer <PB-JWT>` to the server and listen to PocketBase realtime feeds for state refresh.
