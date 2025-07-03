### Project State Update (2024-07-03)

**Today's work:**

- Fully migrated backend from PocketBase to PostgreSQL (persistent) and Redis (ephemeral/session).
- Implemented user registration and login with JWT authentication (passwords hashed with bcrypt, users table in Postgres).
- All game/session logic, move logging, and state are now handled by the new stack.
- Added .env loading for local dev.
- Implemented game archival: finished games are saved in Postgres and become immutable.
- All PocketBase code and references removed.
- End-to-end flow: register/login → create/join game → play → finish → game is archived and cannot be modified.

### OK What the **current foundation already covers**

| Component | Status |
|-----------|--------|
| **PostgreSQL/Redis stack** | All game, session, move, and user data handled by new stack |
| **Session management**  | • `GET /game/{id}` returns sanitized view per player<br>• `POST /game/{id}/join` with spectator mode<br>• `POST /game/{id}/start` with deck building & dealing<br>• All in-game actions (moves, ready, etc.) handled via WebSocket |
| **Move validation**     | • Turn order enforcement<br>• Cost deduction (gems/coins)<br>• Power order chain (0 -> 1)<br>• Challenge window (5s) with rollback |
| **Power engine**        | • 26 active powers implemented<br>• 3 passive powers (alternate_strength, teamwork, keep_gems)<br>• Action registry with dependency injection |
| **Authentication**      | • JWT verification middleware<br>• Register/login with bcrypt password hashing<br>• User ID injected into request context |
| **Game archival**       | • Finished games are archived in Postgres and become immutable |
| **Session lifecycle**   | • `POST /game` to create session (deck building, respecting `default_amount_per_session`) <br>• Optional `/game/{id}/forfeit` |
| **Moves logging**       | • Write each validated move to `moves` table for audit/replay |
| **Real-time updates**   | • All in-game actions and state changes are pushed via WebSocket |
| **Status handling**     | • Update `games_accounts.status` (`Online`, `Lost connection`, ...) on connect/disconnect |
| **Validation & concurrency** | • Input validation <br>• Locking / optimistic concurrency to avoid race conditions |
| **Testing**             | • Manual and Postman-based integration tests for full cycle |
| **Admin / tools**       | • Optional CLI or script to seed cards & powers |
| **Observability**       | • Structured logging, error wrapping <br>• Health-check endpoint |
| **Deployment polish**   | • Coolify service definition / env secrets <br>• CI pipeline (lint, vet, test, Docker build) |

### **Remaining work**

| Area | To-do |
|------|-------|
| **Game logic core**          | • Enforce power **order** chain <br>• Deduct **costs** (gems/coins) <br>• Execute **actions** (`reveal_role`, `steal_gems`, etc.) <br>• Handle **passive** triggers (`trigger` field) <br>• Respect `strength`, combat, loot distribution <br>• Win conditions (glory = 4, last survivor) |
| **Testing**                  | • Unit tests for `services/game.go` <br>• Integration tests for full cycle |
| **Admin / tools**            | • Optional CLI or script to seed cards & powers |
| **Docs**                     | • Public API schema (OpenAPI or MD) for client app |
| **Frontend polish**          | • UI integration with new backend |

The foundation is now fully on PostgreSQL/Redis, with real auth and game archival. Next steps focus on polish, testing, and frontend integration.
