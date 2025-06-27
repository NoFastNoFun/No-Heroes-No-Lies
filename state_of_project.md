### ✅ What the **current foundation already covers**

| Area                    | Implemented Details                                                                                                                    |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **Project scaffolding** | Go module (`go.mod`), folder structure under `cmd/` and `internal/`                                                                    |
| **Config loader**       | Reads `PORT`, `POCKETBASE_URL`, `POCKETBASE_TOKEN`, `POCKETBASE_AO_KEY`                                                                |
| **PocketBase client**   | Re-usable REST helper with:<br>• `Authorization` header<br>• mandatory `AO_KEY` header<br>• `FetchSession` and `UpdateSession` methods |
| **Data models**         | Strongly-typed structs for `Power`, `Card`, `PlayerState`, `GameState`, `GameSession`                                                  |
| **HTTP server**         | Chi router, `/game/{id}` **GET** (fetch session) and `/game/{id}/move` **POST** (submit move)                                          |
| **Game service**        | `FetchSession` wrapper and minimal `ApplyMove` (turn-order check + advance turn)                                                       |
| **Docker image**        | Two-stage build, static binary, Alpine runtime                                                                                         |
| **PocketBase schema**   | Collections: `games_accounts`, `powers`, `cards`, `game_sessions`, `moves` (with `order` in powers, `power_ids` in cards, etc.)        |
| **Security rule hook**  | AO\_KEY header requirement satisfied globally                                                                                          |
| **Naming / repo**       | Project follows **No-Heroes-No-Lies** naming convention                                                                                |

---

### 🚧 **Remaining work**

| Category                     | Tasks still to do                                                                                                                                                                                                                                                                         |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Authentication**           | • JWT verification middleware<br>• Inject `playerID` into request context                                                                                                                                                                                                                 |
| **Game logic core**          | • Enforce power **order** chain <br>• Deduct **costs** (gems/coins) <br>• Execute **actions** (`reveal_role`, `steal_gems`, etc.) <br>• Handle **passive** triggers (`trigger` field) <br>• Respect `strength`, combat, loot distribution <br>• Win conditions (glory = 4, last survivor) |
| **Session lifecycle**        | • `POST /game` to create session (deck building, respecting `default_amount_per_session`) <br>• Optional `/game/{id}/forfeit`                                                                                                                                                             |
| **Moves logging**            | • Write each validated move to `moves` collection for audit/replay                                                                                                                                                                                                                        |
| **Real-time updates**        | • Broadcast state changes (PocketBase subscription or WebSocket push)                                                                                                                                                                                                                     |
| **Status handling**          | • Update `games_accounts.status` (`Online`, `Lost connection`, …) on connect/disconnect                                                                                                                                                                                                   |
| **Validation & concurrency** | • Input validation <br>• Locking / optimistic concurrency to avoid race conditions                                                                                                                                                                                                        |
| **Testing**                  | • Unit tests for `services/game.go` <br>• Integration tests against PocketBase mock                                                                                                                                                                                                       |
| **Admin / tools**            | • Optional CLI or script to seed cards & powers                                                                                                                                                                                                                                           |
| **Observability**            | • Structured logging, error wrapping <br>• Health-check endpoint                                                                                                                                                                                                                          |
| **Deployment polish**        | • Coolify service definition / env secrets <br>• CI pipeline (lint, vet, test, Docker build)                                                                                                                                                                                              |

This checklist keeps the project on track: the **foundation is in place**, and the next milestones focus on game-rule implementation, authentication, and real-time interactions.
