# No-Heroes-No-Lies - PocketBase Collections

---

## games_accounts

| Field        | Type            | Notes                                           |
|--------------|-----------------|-------------------------------------------------|
| id           | string (UUID)   | Primary key                                     |
| email        | string          | Unique                                          |
| username     | string          | Unique handle                                   |
| password     | string (hashed) | PocketBase auth                                 |
| status       | enum            | `Online` \| `Lost connection` \| `Away` \| `Offline` |
| display_name | string          | Name shown in game                              |
| avatar       | file            | Player avatar image                             |
| created_at   | timestamp       | Auto-generated                                  |
| updated_at   | timestamp       | Auto-generated                                  |

---

## powers

| Field      | Type            | Notes                                                        |
|------------|-----------------|--------------------------------------------------------------|
| id         | string (UUID)   | Primary key                                                 |
| name       | string          | Power title                                                 |
| type       | enum            | `active` \| `passive`                                       |
| cost       | int             | Gems / coins spent (0 for passive or free)                  |
| action     | string          | Backend keyword (e.g. `reveal_role`, `steal_gems`)          |
| target     | string          | `self`, `other_player`, `session`, etc.                     |
| trigger    | string?         | Passive only (`on_steal_attempt`, `on_monster_slain`, …)    |
| description| string          | Optional player-facing text                                 |
| created_at | timestamp       | Auto-generated                                              |
| updated_at | timestamp       | Auto-generated                                              |

---

## cards

| Field                      | Type            | Notes                                                    |
|----------------------------|-----------------|----------------------------------------------------------|
| id                         | string (UUID)   | Primary key                                             |
| type                       | string          | `hero`, `monster`, etc.                                 |
| name                       | string          | Card title                                              |
| description                | string          | Player-facing text                                      |
| strength                   | int             | Combat value                                            |
| power_ids                  | list\<id>       | Relations → `powers` collection                          |
| loot                       | JSON            | `{ "coins": int, "gems": int }` (monsters)              |
| default_amount_per_session | int             | Max copies of this card per game session                |
| cover                      | file            | Card image stored in PocketBase                         |
| created_at                 | timestamp       | Auto-generated                                          |
| updated_at                 | timestamp       | Auto-generated                                          |

---

## game_sessions

| Field       | Type            | Notes                                      |
|-------------|-----------------|--------------------------------------------|
| id          | string (UUID)   | Primary key                                |
| player_ids  | list\<id>       | Relations → `games_accounts`               |
| state       | JSON            | Authoritative game data (see template)     |
| is_active   | bool            | Game in progress?                          |
| created_at  | timestamp       | Auto-generated                             |
| updated_at  | timestamp       | Auto-generated                             |

### state JSON template

```json
{
  "players": [
    {
      "id": "player-id",
      "life": 2,
      "coins": 0,
      "gems": 0,
      "glory": 0,
      "current_hero": "hero-card-id",
      "current_alibi": "claimed-hero-id"
    }
  ],
  "turn_order": ["player-id-1", "player-id-2"],
  "current_turn": "player-id-2",
  "coin_pool": 20,
  "gem_pool": 10
}
````

---

## moves (optional)

| Field       | Type          | Notes                       |
| ----------- | ------------- | --------------------------- |
| id          | string (UUID) | Primary key                 |
| session\_id | id            | Relation → `game_sessions`  |
| player\_id  | id            | Relation → `games_accounts` |
| move\_data  | JSON          | Action payload              |
| created\_at | timestamp     | Auto-generated              |
| updated\_at | timestamp     | Auto-generated              |
