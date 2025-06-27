# Heroes Love to Lie - Custom Game Server

This server enforces the game rules and maintains authoritative game state for the multiplayer card game **"Heroes Love to Lie"**, ensuring fairness, validation, and persistence.

It is designed to work alongside **PocketBase** (which handles Auth, database, and real-time subscriptions) and is intended to be deployed via **Coolify** on our existing VPS.

---

## 🎯 Purpose

✅ Provide a secure, authoritative layer to validate and process player moves  
✅ Enforce game rules, turns, and hidden roles/bluffing logic  
✅ Update game session state in PocketBase  
✅ Expose a simple HTTP API for the client to interact with  

---

## ⚙️ Architecture Assumptions

- **PocketBase** handles:
  - User accounts and authentication
  - Persistent storage of cards, game sessions, moves
  - Real-time subscriptions for clients to receive state updates

- **This server**:
  - Accepts player moves
  - Validates legality of moves
  - Updates PocketBase game session state
  - Keeps the game logic centralized and tamper-proof

---

## ✅ Feature List / TODO

### ✅ 1. Auth
- [ ] Verify PocketBase user tokens via API
- [ ] Middleware to attach authenticated user to request

---

### ✅ 2. Data Models (PocketBase Collections)
- [ ] `users`
  - id, email, password, display_name
- [ ] `cards`
  - id, type (hero/monster), name, description, powers (JSON)
- [ ] `game_sessions`
  - id, player_ids (list of user ids), state (JSON), is_active (bool)
- [ ] `moves` *(optional)*
  - id, session_id, player_id, move_data (JSON)

---

### ✅ 3. API Endpoints
- [ ] `POST /game`
  - Create a new game session
  - Assign players
  - Seed initial game state

- [ ] `GET /game/:id`
  - Fetch current game session state

- [ ] `POST /game/:id/move`
  - Receive and validate a player's move
  - Enforce turn order
  - Enforce hidden role/bluffing rules
  - Update state in PocketBase

- [ ] `POST /game/:id/forfeit`
  - Handle player surrender or leaving the game

---

### ✅ 4. Game Logic
- [ ] Validate moves against current state
- [ ] Enforce turn order
- [ ] Handle hidden roles and bluffing rules
- [ ] Reject invalid or cheating moves
- [ ] Transition game state correctly

---

### ✅ 5. PocketBase Integration
- [ ] Connect to PocketBase via REST or SDK
- [ ] CRUD operations for:
  - Game sessions
  - Moves
  - Cards
- [ ] Support admin seeding of cards/heroes

---

### ✅ 6. Real-Time Support
- [ ] Optionally trigger PocketBase subscriptions on session updates
- [ ] Ensure minimal latency for turn-based play

---

### ✅ 7. Deployment
- [ ] Create Dockerfile
- [ ] Set up environment variables (PocketBase URL, API key)
- [ ] Deploy via Coolify on VPS

---

### ✅ 8. Testing
- [ ] Unit tests for move validation logic
- [ ] Integration tests for PocketBase updates

---

## 💡 Notes

- This server does **not** handle static file serving or frontend.  
- Client apps will communicate via HTTP API, and listen to PocketBase real-time updates for session state.  
- Admins can add/edit cards and heroes directly in PocketBase’s built-in Admin UI.
