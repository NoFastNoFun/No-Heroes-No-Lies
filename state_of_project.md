### OK What the **current foundation already covers**

| Component | Status |
|-----------|--------|
| **PocketBase client**   | Re-usable REST helper with:<br>• `Authorization` header<br>• mandatory `AO_KEY` header<br>• `FetchSession` and `UpdateSession` methods |
| **Session management**  | • `GET /game/{id}` returns sanitized view per player<br>• `POST /game/{id}/join` with spectator mode<br>• `POST /game/{id}/start` with deck building & dealing<br>• `POST /game/{id}/move` with validation & persistence |
| **Move validation**     | • Turn order enforcement<br>• Cost deduction (gems/coins)<br>• Power order chain (0 -> 1)<br>• Challenge window (5s) with rollback |
| **Power engine**        | • 26 active powers implemented<br>• 3 passive powers (alternate_strength, teamwork, keep_gems)<br>• Action registry with dependency injection |

### **Remaining work**

| Area | To-do |
|------|-------|
| **Authentication**           | • JWT verification middleware<br>• Inject `playerID` into request context
| **Game logic core**          | • Enforce power **order** chain <br>• Deduct **costs** (gems/coins) <br>• Execute **actions** (`reveal_role`, `steal_gems`, etc.) <br>• Handle **passive** triggers (`trigger` field) <br>• Respect `strength`, combat, loot distribution <br>• Win conditions (glory = 4, last survivor) |
| **Session lifecycle**        | • `POST /game` to create session (deck building, respecting `default_amount_per_session`) <br>• Optional `/game/{id}/forfeit`                                                                                                                                                             |
| **Moves logging**            | • Write each validated move to `moves` collection for audit/replay
| **Real-time updates**        | • Broadcast state changes (PocketBase subscription or WebSocket push)
| **Status handling**          | • Update `games_accounts.status` (`Online`, `Lost connection`, ...) on connect/disconnect
| **Validation & concurrency** | • Input validation <br>• Locking / optimistic concurrency to avoid race conditions
| **Testing**                  | • Unit tests for `services/game.go` <br>• Integration tests against PocketBase mock
| **Admin / tools**            | • Optional CLI or script to seed cards & powers
| **Observability**            | • Structured logging, error wrapping <br>• Health-check endpoint
| **Deployment polish**        | • Coolify service definition / env secrets <br>• CI pipeline (lint, vet, test, Docker build)

This checklist keeps the project on track: the **foundation is in place**, and the next milestones focus on game-rule implementation, authentication, and real-time interactions.
