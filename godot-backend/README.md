# No Heroes No Lies - Godot Backend

A complete Godot Engine backend implementation of the "No Heroes No Lies" game, converted from the original Go backend. This implementation preserves all features and game mechanics while providing enhanced development experience and real-time data loading.

## 🎮 Complete Feature Set

### Core Game Mechanics

- **Session Management**: Create, join, and manage game sessions
- **Player Management**: Add/remove players, spectator mode, ready system
- **Turn-based Gameplay**: Proper turn order and validation
- **Challenge System**: 5-second challenge windows for moves
- **Win Conditions**: Multiple victory conditions (last player standing, coin threshold)
- **Draw Conditions**: Stalemate detection when all players die

### Card Browser System

The Godot backend includes a comprehensive **Card Browser** that provides:

- **Visual Card Display**: Browse all cards with images and details
- **Category Organization**: Separate views for Heroes and Monsters
- **Strength Sorting**: Cards sorted by strength (strongest to weakest)
- **Tabbed Interface**:
  - "All Cards" - Combined view with both categories
  - "Heroes Only" - Hero cards only
  - "Monsters Only" - Monster cards only
- **Card Details Panel**: Click any card to view:
  - Card image and artwork
  - Name, type, and strength
  - Full description
  - Associated powers with descriptions
- **Search and Filter**: Find cards by name or description
- **Real-time Updates**: Automatically refreshes when new cards are loaded
- **Interactive UI**: Hover effects and click interactions

### Advanced Power System

The Godot backend implements the complete power system from the Go backend with **20+ different powers**:

#### Active Powers

- `change_monster` - Swap active monster with top of deck
- `fight_monster` - Free monster fight flag
- `fight_player_with_discarded_card` - Strength duel with discarded hero
- `add_strength_to_challenger` - +1 strength for same-turn fights
- `shuffle_heroes_deck` - Recycle discard pile and shuffle
- `change_hero` - Draw new hero, discard current
- `exchange_gems_with_player` - Swap gems with target player
- `exchange_2coins_for_life` - Trade 2 coins for 1 life
- `gain_2gems` - Gain 2 gems (triggers events)
- `gain_life` - Gain 1 life point
- `see_player_card` - View target's current hero
- `remove_coin` - Remove 1 coin from target
- `shoot_player` - Attack with hero guess
- `all_in` - Double gems if last shot was successful
- `steal_gem_0/1` - Steal gems (order 0 and 1)
- `parry_this_you_f_casual` - Direct life damage
- `steal_coin` - Steal coins from target
- `mimic_power` - Copy target's active power
- `dual_attack` - Enable dual attack mode
- `blind_draw` - Force target to draw new hero
- `force_transform` - Transform target's hero
- `execution` - Permanently remove hero cards

#### Passive Powers (Trigger System)

- `keep_gems` - Cancel gem theft attempts
- `tribute` - Receive half of gems gained by others
- `teamwork` - Duplicate gems when others gain them

### Event-Driven Trigger System

The backend implements a sophisticated event system for passive abilities:

- **steal_attempt** - Triggered when gems are stolen
- **gems_gained** - Triggered when players gain gems
- Extensible event system for future powers

### Card Management

- **Hero Cards**: Player identities with powers and strength
- **Monster Cards**: Combat encounters with loot rewards
- **Deck Management**: Shuffle, draw, discard, and recycle mechanics
- **Card Execution**: Permanent removal of cards from game
- **Burned Cards**: Hidden cards removed from memory
- **Public Discard**: Last discarded card (public knowledge)

### Combat System

- **Strength Calculation**: Base strength + bonus strength
- **Monster Combat**: Hero vs monster with loot rewards
- **Player Combat**: Hero vs hero duels
- **Life System**: Player health with death mechanics
- **Loot System**: Coins and gems from victories

### Advanced Game Features

- **Dual Attack Mode**: Special combat state
- **Bonus Strength**: Temporary strength modifiers
- **Alibi System**: Declared vs actual hero identities
- **Bluffing Mechanics**: Deception and challenge system
- **Spectator Mode**: Read-only game observation
- **Forfeit System**: Player surrender mechanics

## 🏗️ Architecture

### Core Components

- **GameLogic**: Complete game engine with power system
- **GameModels**: Data structures matching Go backend
- **Database**: PocketBase integration for data persistence
- **GameServer**: HTTP server with API endpoints
- **AuthManager**: JWT authentication system
- **ConfigManager**: Configuration management
- **CardBrowser**: Visual card browsing interface

### Power System Architecture

```
Power Registry → Effect Functions → Game State Updates
     ↓
Trigger System → Event Handlers → Passive Abilities
     ↓
Game Logic → Win Detection → Session Management
```

### Card Browser Architecture

```
CardBrowser → CardWidget → Card Details Panel
     ↓
GameLogic.get_all_cards() → PocketBase → Real-time Data
     ↓
Category Filtering → Strength Sorting → Visual Display
```

## 🚀 Getting Started

### Prerequisites

- Godot 4.x
- PocketBase server running
- Cards and Powers collections configured

### Installation

1. Clone the repository
2. Open `project.godot` in Godot
3. Configure `config.json` with your PocketBase settings
4. Run the server scene

### Configuration

Update `config.json` with your PocketBase settings:

```json
{
  "pocketbase_url": "http://localhost:8090",
  "pocketbase_admin_email": "admin@example.com",
  "pocketbase_admin_password": "your_password",
  "server_port": 8080,
  "jwt_secret": "your_jwt_secret"
}
```

## 🎮 Using the Card Browser

### Opening the Card Browser

1. Start the server application
2. Click the "Open Card Browser" button in the main UI
3. Or press **F5** to toggle the card browser

### Features

- **Browse by Category**: Switch between "All Cards", "Heroes Only", and "Monsters Only" tabs
- **View Card Details**: Click any card to see full information including powers
- **Refresh Data**: Click "Refresh Cards" to reload from PocketBase
- **Search Cards**: Use the search functionality to find specific cards
- **Sort by Strength**: Cards are automatically sorted by strength (strongest first)

### Card Information Displayed

- **Card Image**: Visual representation (placeholder or actual artwork)
- **Name**: Card name
- **Type**: Hero or Monster
- **Strength**: Combat strength value
- **Description**: Full card description
- **Powers**: List of associated powers with descriptions

## 📊 PocketBase Collections

### Cards Collection

```json
{
  "id": "string",
  "name": "string",
  "type": "hero|monster",
  "strength": "number",
  "power_ids": ["string"],
  "loot": {
    "coins": "number",
    "gems": "number"
  },
  "default_amount_per_session": "number",
  "description": "string",
  "cover_url": "string"
}
```

### Powers Collection

```json
{
  "id": "string",
  "name": "string",
  "type": "active|passive",
  "cost": "number",
  "action": "string",
  "target": "string",
  "trigger": "string",
  "order": "number",
  "description": "string"
}
```

## 🔧 API Endpoints

### Authentication

- `POST /api/auth/login` - Player login
- `POST /api/auth/register` - Player registration
- `GET /api/auth/verify` - Verify JWT token

### Game Sessions

- `POST /api/game` - Create new session
- `GET /api/game/{id}` - Get session state
- `POST /api/game/{id}/join` - Join session
- `POST /api/game/{id}/start` - Start game
- `POST /api/game/{id}/ready` - Toggle ready status
- `POST /api/game/{id}/move` - Submit move
- `POST /api/game/{id}/forfeit` - Forfeit game

### Game Data

- `GET /api/cards` - Get all cards
- `GET /api/cards/{id}` - Get specific card
- `GET /api/powers` - Get all powers
- `GET /api/powers/{id}` - Get specific power

### Health

- `GET /api/health` - Health check

## 🎯 Game Flow

1. **Session Creation**: Players create or join a session
2. **Lobby Phase**: Players ready up, session starts when all ready
3. **Game Initialization**: Cards dealt, monsters drawn, turn order set
4. **Gameplay Loop**:
   - Player takes turn (demask/fight/power)
   - 5-second challenge window
   - Next player's turn
   - Win condition check
5. **Game End**: Victory or draw condition met

## 🔄 Migration from Go Backend

This Godot implementation preserves 100% of the original Go backend functionality:

### Preserved Features

- ✅ All 20+ powers and their effects
- ✅ Complete trigger system with passive abilities
- ✅ Win/draw condition detection
- ✅ Challenge system with timeouts
- ✅ Card execution and burning mechanics
- ✅ Dual attack and bonus strength systems
- ✅ Alibi and bluffing mechanics
- ✅ Spectator and forfeit systems
- ✅ PocketBase integration
- ✅ JWT authentication
- ✅ CORS support
- ✅ API compatibility

### Enhanced Features

- 🚀 Real-time data loading from PocketBase
- 🎮 Better development experience with Godot
- 📊 Enhanced debugging and visualization
- 🔧 Easier configuration management
- 📱 Better cross-platform support

## 🧪 Testing

### Test PocketBase Integration

Run the test script to verify data loading:

```gdscript
# In Godot editor, run TestPocketBase scene
# Or use the test script directly
```

### Test Game Logic

```gdscript
# Create test game state
var game_logic = GameLogic.new()
var state = game_logic.initialize_game_session(["player1", "player2"])

# Test power execution
var move = GameModels.MovePayload.new({
    "type": "power",
    "power_id": "gain_2gems"
})
var result = game_logic.process_move(state, "player1", move)
```

## 📝 Development Notes

### Power System Design

The power system uses a registry pattern where each power action is registered with its effect function. This allows for easy extension and modification of powers without changing core game logic.

### Event System

The trigger system uses events to handle passive abilities. Events are dispatched when certain actions occur (like gaining gems), and registered handlers can modify the game state or cancel the action.

### Win Conditions

Multiple victory conditions are checked after each move:

1. **Last Player Standing**: Only one player with life > 0
2. **Coin Victory**: Player(s) with 5+ coins
3. **Draw**: All players dead simultaneously

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This project maintains the same license as the original Go backend.

---

**Status**: ✅ Complete - All Go backend features implemented and tested
