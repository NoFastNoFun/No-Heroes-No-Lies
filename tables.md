# PostgreSQL Tables Schema (Post-PocketBase Migration)

## users

| Field         | Type    | Description                |
|-------------- |---------|----------------------------|
| id            | UUID    | Primary key                |
| email         | string  | Unique email               |
| username      | string  | Unique player name         |
| password_hash | string  | Bcrypt hash                |
| created_at    | datetime| Registration timestamp     |

## cards

| Field                     | Type    | Description                  |
|---------------------------|---------|------------------------------|
| id                        | string  | Primary key                  |
| name                      | string  | Card name                    |
| strength                  | int     | Combat value                 |
| default_spawn             | int     | How many copies in deck      |
| rarity                    | string  | Rarity                       |
| power1, power2            | string  | Power names                  |
| notes                     | string  | Card notes                   |

## powers

| Field        | Type    | Description                  |
|--------------|---------|------------------------------|
| name         | string  | Primary key                  |
| cost         | int     | Gems required                |
| target       | string  | Target type                  |
| description  | string  | What it does                 |
| is_passive   | bool    | Passive power?               |
| trigger      | string  | Passive trigger event        |

## game_sessions (in Redis)

| Field         | Type    | Description                  |
|---------------|---------|------------------------------|
| id            | string  | Primary key                  |
| player_ids    | string[]| Player IDs                   |
| state         | JSON    | GameState struct             |
| is_active     | bool    | Game started                 |
| created_at    | datetime| Creation timestamp           |
| updated_at    | datetime| Last update                  |

## moves

| Field         | Type    | Description                  |
|---------------|---------|------------------------------|
| id            | string  | Primary key                  |
| session_id    | string  | Game session ID              |
| player_id     | string  | Player ID                    |
| type          | string  | Move type                    |
| move_data     | string  | JSON payload                 |
| created_at    | datetime| Timestamp                    |

## archived_games

| Field         | Type    | Description                  |
|---------------|---------|------------------------------|
| id            | string  | Primary key                  |
| player_ids    | string[]| Player IDs                   |
| state         | JSON    | Final GameState              |
| moves         | JSON    | All moves (for replay/audit) |
| is_active     | bool    | Should be false (archived)   |
| created_at    | datetime| Creation timestamp           |
| updated_at    | datetime| Last update                  |

---

**Note:**

- All game/session/move logic is now handled by PostgreSQL and Redis.
- PocketBase is no longer used.
- User authentication is via the `users` table and JWT.
- Finished games are archived and immutable.
