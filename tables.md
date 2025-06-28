# PocketBase Collections Schema

## games_accounts

| Field | Type | Description |
|-------|------|-------------|
| id | id | Primary key |
| created | datetime | Auto-generated |
| updated | datetime | Auto-generated |
| username | string | Unique player name |
| status | string | `Online`, `Offline`, `Lost connection` |
| last_seen | datetime | Last activity timestamp |

## powers

| Field | Type | Description |
|-------|------|-------------|
| id | id | Primary key |
| created | datetime | Auto-generated |
| updated | datetime | Auto-generated |
| name | string | Human-readable name |
| description | string | What it does |
| action | string | Function name to call |
| type | string | `active` or `passive` |
| cost | number | Gems required |
| order | number | 0 or 1 (chain requirement) |
| trigger | string? | Passive only (`on_steal_attempt`, `on_monster_slain`, ...) |

## cards

| Field | Type | Description |
|-------|------|-------------|
| id | id | Primary key |
| created | datetime | Auto-generated |
| updated | datetime | Auto-generated |
| name | string | Card name |
| type | string | `hero` or `monster` |
| strength | number | Combat value |
| description | string | Flavor text |
| default_amount_per_session | number | How many copies in deck |
| power_ids | list<id> | Relations -> `powers` collection |
| loot_coins | number? | Monster only |
| loot_gems | number? | Monster only |

## game_sessions

| Field | Type | Description |
|-------|------|-------------|
| id | id | Primary key |
| created | datetime | Auto-generated |
| updated | datetime | Auto-generated |
| player_ids | list<id> | Relations -> `games_accounts` |
| spectator_ids | list<id> | Relations -> `games_accounts` |
| ready_ids | list<id> | Relations -> `games_accounts` |
| is_active | bool | Game started |
| state | json | GameState struct |
| seed | number | RNG seed for replay |

## moves

| Field | Type | Description |
|-------|------|-------------|
| id | id | Primary key |
| created | datetime | Auto-generated |
| updated | datetime | Auto-generated |
| session_id | id | Relation -> `game_sessions` |
| player_id | id | Relation -> `games_accounts` |
| type | string | `demask`, `fight`, `power` |
| data | string | JSON payload |
