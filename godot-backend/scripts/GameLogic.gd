extends RefCounted

# Game logic engine that handles core game mechanics
# Complete implementation matching the Go backend

class_name GameLogic

# Game constants
const DEFAULT_LIFE = 3
const DEFAULT_COINS = 0
const DEFAULT_GEMS = 0
const DEFAULT_GLORY = 0
const COIN_WIN_THRESHOLD = 5  # Win condition: 5 coins

# Game state management
var cards_data: Dictionary = {}
var powers_data: Dictionary = {}
var data_loaded: bool = false

# Power registry - maps action names to effect functions
var power_registry: Dictionary = {}

# Trigger registry - maps event types to handler functions
var trigger_registry: Dictionary = {}

signal data_loaded_signal
signal cards_loaded(cards: Array)
signal powers_loaded(powers: Array)
signal power_executed(power_id: String, actor_id: String, target_id: String, success: bool)

# Event structure for triggers
class Event:
	var type: String
	var actor_id: String
	var target_id: String
	var amount: int
	var meta: Dictionary
	var cancel: bool
	
	func _init(event_type: String, actor: String, target: String = "", amt: int = 0, metadata: Dictionary = {}):
		type = event_type
		actor_id = actor
		target_id = target
		amount = amt
		meta = metadata
		cancel = false

func _init():
	# Connect to database signals
	Database.cards_loaded.connect(_on_cards_loaded)
	Database.powers_loaded.connect(_on_powers_loaded)
	
	# Load game data from PocketBase
	load_game_data()
	
	# Initialize power system
	_register_all_powers()
	_register_all_triggers()

func load_game_data():
	# Load cards and powers from PocketBase
	Database.get_cards()
	Database.get_powers()

func _on_cards_loaded(cards: Array):
	# Convert cards array to dictionary for easy lookup
	for card in cards:
		cards_data[card.id] = card
	
	print("Loaded " + str(cards.size()) + " cards from PocketBase")
	cards_loaded.emit(cards)
	_check_data_loaded()

func _on_powers_loaded(powers: Array):
	# Convert powers array to dictionary for easy lookup
	for power in powers:
		powers_data[power.id] = power
	
	print("Loaded " + str(powers.size()) + " powers from PocketBase")
	powers_loaded.emit(powers)
	_check_data_loaded()

func _check_data_loaded():
	if cards_data.size() > 0 and powers_data.size() > 0:
		data_loaded = true
		data_loaded_signal.emit()
		print("All game data loaded successfully")

# Wait for data to be loaded before proceeding
func wait_for_data():
	if not data_loaded:
		await data_loaded_signal

# Register all active powers from the Go backend
func _register_all_powers():
	_register_power("change_monster", _change_monster_effect)
	_register_power("fight_monster", _fight_monster_effect)
	_register_power("fight_player_with_discarded_card", _fight_player_with_discarded_card_effect)
	_register_power("add_strength_to_challenger", _add_strength_to_challenger_effect)
	_register_power("shuffle_heroes_deck", _shuffle_heroes_deck_effect)
	_register_power("change_hero", _change_hero_effect)
	_register_power("exchange_gems_with_player", _exchange_gems_with_player_effect)
	_register_power("exchange_2coins_for_life", _exchange_2coins_for_life_effect)
	_register_power("gain_2gems", _gain_2gems_effect)
	_register_power("gain_life", _gain_life_effect)
	_register_power("see_player_card", _see_player_card_effect)
	_register_power("remove_coin", _remove_coin_effect)
	_register_power("shoot_player", _shoot_player_effect)
	_register_power("all_in", _all_in_effect)
	_register_power("steal_gem_0", _steal_gem_effect)
	_register_power("steal_gem_1", _steal_gem_effect)
	_register_power("parry_this_you_f_casual", _parry_this_you_f_casual_effect)
	_register_power("steal_coin", _steal_coin_effect)
	_register_power("mimic_power", _mimic_power_effect)
	_register_power("dual_attack", _dual_attack_effect)
	_register_power("blind_draw", _blind_draw_effect)
	_register_power("force_transform", _force_transform_effect)
	_register_power("execution", _execution_effect)

# Register all trigger handlers
func _register_all_triggers():
	_register_trigger("steal_attempt", _keep_gems_trigger)
	_register_trigger("gems_gained", _tribute_trigger)
	_register_trigger("gems_gained", _teamwork_trigger)

# Register a power effect
func _register_power(action: String, effect_func: Callable):
	power_registry[action] = effect_func

# Register a trigger handler
func _register_trigger(event_type: String, handler_func: Callable):
	if not trigger_registry.has(event_type):
		trigger_registry[event_type] = []
	trigger_registry[event_type].append(handler_func)

# Execute a power
func execute_power(state: GameModels.GameState, actor: GameModels.PlayerState, power: Dictionary, move: GameModels.MovePayload) -> Dictionary:
	var action = power.get("action", "")
	
	if not power_registry.has(action):
		return {"success": false, "error": "Unknown power action: " + action}
	
	var effect_func = power_registry[action]
	var result = effect_func.call(state, actor, move)
	
	power_executed.emit(power.get("id", ""), actor.id, move.target_player, result.get("success", false))
	
	return result

# Dispatch an event to trigger handlers
func dispatch_event(state: GameModels.GameState, event: Event) -> bool:
	if not trigger_registry.has(event.type):
		return false
	
	var handlers = trigger_registry[event.type]
	for handler in handlers:
		handler.call(state, event)
		if event.cancel:
			return true
	
	return false

# Initialize a new game session
func initialize_game_session(player_ids: Array[String]) -> GameModels.GameState:
	# Wait for data to be loaded
	await wait_for_data()
	
	var state = GameModels.GameState.new()
	
	# Add players
	for player_id in player_ids:
		var player = GameModels.PlayerState.new({
			"id": player_id,
			"life": DEFAULT_LIFE,
			"coins": DEFAULT_COINS,
			"gems": DEFAULT_GEMS,
			"glory": DEFAULT_GLORY
		})
		state.add_player(player)
	
	# Initialize decks
	state.hero_deck = _create_hero_deck()
	state.monster_deck = _create_monster_deck()
	
	# Shuffle decks
	state.hero_deck.shuffle()
	state.monster_deck.shuffle()
	
	# Deal initial hands
	_deal_initial_hands(state)
	
	return state

func _create_hero_deck() -> Array[String]:
	var deck = []
	for card_id in cards_data:
		var card = cards_data[card_id]
		if card.type == "hero":
			# Add cards based on default_amount_per_session
			var amount = card.get("default_amount_per_session", 1)
			for i in range(amount):
				deck.append(card_id)
	return deck

func _create_monster_deck() -> Array[String]:
	var deck = []
	for card_id in cards_data:
		var card = cards_data[card_id]
		if card.type == "monster":
			# Add cards based on default_amount_per_session
			var amount = card.get("default_amount_per_session", 1)
			for i in range(amount):
				deck.append(card_id)
	return deck

func _deal_initial_hands(state: GameModels.GameState):
	for player in state.players:
		# Deal 3 hero cards to each player
		for i in range(3):
			if state.hero_deck.size() > 0:
				var card = state.hero_deck.pop_back()
				player.hand.append(card)

# Process a move submitted by a player
func process_move(state: GameModels.GameState, player_id: String, move: GameModels.MovePayload) -> Dictionary:
	# Wait for data to be loaded
	await wait_for_data()
	
	var player = state.get_player_by_id(player_id)
	if not player:
		return {"success": false, "error": "Player not found"}
	
	if not _is_player_turn(state, player_id):
		return {"success": false, "error": "Not your turn"}
	
	if not move.is_valid():
		return {"success": false, "error": "Invalid move"}
	
	match move.type:
		"demask":
			return _process_demask_move(state, player, move)
		"fight":
			return _process_fight_move(state, player, move)
		"power":
			return _process_power_move(state, player, move)
		_:
			return {"success": false, "error": "Unknown move type"}

# Process demask move (from Go backend)
func _process_demask_move(state: GameModels.GameState, player: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target_player = state.get_player_by_id(move.target_player)
	if not target_player:
		return {"success": false, "error": "Target player not found"}
	
	if player.gems < 6:
		return {"success": false, "error": "Not enough gems"}
	
	player.gems -= 6
	
	if target_player.current_hero == move.guess_hero:
		# Demask successful - target loses life and must change hero
		target_player.life -= 1
		
		# Force target to draw a new hero card
		if state.hero_deck.size() > 0:
			var new_hero = state.hero_deck[0]
			state.hero_deck.remove_at(0)
			_change_player_hero(state, target_player, new_hero)
		else:
			# If no cards in deck, try to recycle discard pile
			_recycle_discard_pile(state)
			
			if state.hero_deck.size() > 0:
				# Successfully recycled, draw new hero
				var new_hero = state.hero_deck[0]
				state.hero_deck.remove_at(0)
				_change_player_hero(state, target_player, new_hero)
			else:
				# No cards available at all, discard current hero and target has no hero
				if target_player.current_hero != "":
					_discard_hero_card(state, target_player.current_hero)
				target_player.current_hero = ""
	
	# Demask has no alibi declaration → no challenge window
	state.next_turn()
	
	# Check for game end conditions
	_detect_outcome(state)
	
	return {
		"success": true,
		"message": "Demask move processed",
		"correct": target_player.current_hero == move.guess_hero
	}

# Process fight move (from Go backend)
func _process_fight_move(state: GameModels.GameState, player: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if move.monster_id == "" or move.declared_alibi == "":
		return {"success": false, "error": "monster_id and declared_alibi required"}
	
	# Check if monster is active
	var active = false
	for monster_id in state.active_monsters:
		if monster_id == move.monster_id:
			active = true
			break
	
	if not active:
		return {"success": false, "error": "Monster not active"}
	
	var monster = cards_data.get(move.monster_id)
	var hero_card = cards_data.get(move.declared_alibi)
	
	if not monster or not hero_card:
		return {"success": false, "error": "Card not found"}
	
	player.current_alibi = move.declared_alibi
	
	var hero_str = _hero_strength(state, player, hero_card)
	var hero_lost = monster.strength > hero_str
	var coins_delta = 0
	var gems_delta = 0
	var life_delta = 0
	
	if hero_lost:
		player.life -= 1
		life_delta = -1
	else:
		# Victory - award loot
		var loot = monster.get("loot", {"coins": 0, "gems": 0})
		coins_delta = loot.get("coins", 0)
		gems_delta = loot.get("gems", 0)
		player.coins += coins_delta
		_give_gems(state, player, gems_delta)
		
		# Remove slain monster
		state.active_monsters.erase(move.monster_id)
		
		# Refill if needed
		if state.active_monsters.size() < 2 and state.monster_deck.size() > 0:
			var next = state.monster_deck[0]
			state.monster_deck.remove_at(0)
			state.active_monsters.append(next)
	
	# Create last move for challenge window
	var last_move = GameModels.LastMove.new({
		"actor_id": player.id,
		"type": "fight",
		"declared_alibi": move.declared_alibi,
		"monster_id": move.monster_id,
		"coins_delta": coins_delta,
		"gems_delta": gems_delta,
		"life_delta_id": player.id if life_delta != 0 else "",
		"life_delta": life_delta,
		"expires_at": Time.get_unix_time_from_system() * 1000 + 5000,  # 5 seconds
		"correct": not hero_lost
	})
	
	state.last_move = last_move
	
	# Move to next turn
	state.next_turn()
	
	# Check for game end conditions
	_detect_outcome(state)
	
	return {
		"success": true,
		"message": "Fight move processed",
		"victory": not hero_lost
	}

# Process power move (from Go backend)
func _process_power_move(state: GameModels.GameState, player: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var power = powers_data.get(move.power_id)
	if not power:
		return {"success": false, "error": "Power not found"}
	
	# Order chain enforcement
	if power.get("order", 0) == 1:
		if not state.order0_used_by.has(player.id) or not state.order0_used_by[player.id]:
			return {"success": false, "error": "must use order-0 power first"}
	
	# Check if player can afford the power
	if player.gems < power.cost:
		return {"success": false, "error": "Not enough gems"}
	
	player.gems -= power.cost
	
	# Apply power effects
	var result = execute_power(state, player, power, move)
	if not result.get("success", false):
		return result
	
	# Mark order 0 usage
	if power.get("order", 0) == 0:
		if not state.order0_used_by.has(player.id):
			state.order0_used_by[player.id] = false
		state.order0_used_by[player.id] = true
	
	# Create last move for challenge window
	var last_move = GameModels.LastMove.new({
		"actor_id": player.id,
		"type": "power",
		"declared_alibi": move.declared_alibi,
		"expires_at": Time.get_unix_time_from_system() * 1000 + 5000,  # 5 seconds
		"correct": true
	})
	
	state.last_move = last_move
	
	# Move to next turn
	state.next_turn()
	
	# Check for game end conditions
	_detect_outcome(state)
	
	return {
		"success": true,
		"message": "Power move processed",
		"power": power.name
	}

# Win condition detection (from Go backend)
func _detect_outcome(state: GameModels.GameState):
	if state.draw or state.winner_ids.size() > 0:
		return  # already ended
	
	var alive: Array[GameModels.PlayerState] = []
	for player in state.players:
		if player.life > 0:
			alive.append(player)
	
	match alive.size():
		0:  # everyone died same turn
			state.draw = true
		1:  # last player standing
			state.winner_ids = [alive[0].id]
		_:  # coin victory check
			var coin_winners: Array[String] = []
			for player in alive:
				if player.coins >= COIN_WIN_THRESHOLD:
					coin_winners.append(player.id)
			if coin_winners.size() > 0:
				state.winner_ids = coin_winners

# Helper functions for game mechanics
func _hero_strength(state: GameModels.GameState, player: GameModels.PlayerState, hero_card: Dictionary) -> int:
	if player.bonus_strength > 0:
		return player.bonus_strength
	return hero_card.get("strength", 0)

func _give_gems(state: GameModels.GameState, player: GameModels.PlayerState, amount: int):
	if amount <= 0:
		return
	
	player.gems += amount
	
	# Trigger gems_gained event
	var event = Event.new("gems_gained", player.id, "", amount)
	dispatch_event(state, event)

func _change_player_hero(state: GameModels.GameState, player: GameModels.PlayerState, new_hero: String):
	# Discard current hero if exists
	if player.current_hero != "":
		_discard_hero_card(state, player.current_hero)
	
	# Set new hero
	player.current_hero = new_hero
	player.current_alibi = ""  # must bluff again

func _discard_hero_card(state: GameModels.GameState, card_id: String):
	if card_id == "":
		return
	state.discard_pile.append(card_id)
	state.public_discard = card_id

func _recycle_discard_pile(state: GameModels.GameState):
	if state.discard_pile.size() == 0:
		return  # nothing to recycle
	
	# Move all non-burned discarded cards back to the deck
	state.hero_deck.append_array(state.discard_pile)
	state.discard_pile.clear()
	state.public_discard = ""  # clear public discard since we're recycling
	
	# Shuffle the recycled deck
	state.hero_deck.shuffle()

func _ensure_hero_deck_not_empty(state: GameModels.GameState) -> bool:
	if state.hero_deck.size() == 0:
		if state.discard_pile.size() > 0:
			state.hero_deck.append_array(state.discard_pile)
			state.discard_pile.clear()
			state.public_discard = ""
			state.hero_deck.shuffle()
			return true
		else:
			return false
	return true

func _has_passive(state: GameModels.GameState, player: GameModels.PlayerState, action: String) -> bool:
	if player.current_alibi == "":
		return false
	
	var card = cards_data.get(player.current_alibi)
	if not card:
		return false
	
	for power_id in card.get("power_ids", []):
		var power = powers_data.get(power_id)
		if power and power.get("type") == "passive" and power.get("action") == action:
			return true
	
	return false

func _is_player_turn(state: GameModels.GameState, player_id: String) -> bool:
	return state.current_turn == player_id

# Start a game session
func start_game(state: GameModels.GameState) -> Dictionary:
	if state.players.size() < 2:
		return {"success": false, "error": "Need at least 2 players"}
	
	# Set first player as current turn
	state.current_turn = state.turn_order[0]
	
	# Clear ready status
	state.ready_ids.clear()
	
	return {"success": true, "message": "Game started"}

# Handle player ready status
func toggle_player_ready(state: GameModels.GameState, player_id: String) -> Dictionary:
	state.toggle_player_ready(player_id)
	
	var ready_count = state.ready_ids.size()
	var total_players = state.players.size()
	
	return {
		"success": true,
		"ready": state.is_player_ready(player_id),
		"ready_count": ready_count,
		"total_players": total_players
	}

# Handle challenge resolution (from Go backend)
func resolve_challenge(state: GameModels.GameState, challenger_id: String, challenge_id: String) -> Dictionary:
	if not state.last_move:
		return {"success": false, "error": "No move to challenge"}
	
	if state.last_move.is_expired():
		return {"success": false, "error": "Challenge window expired"}
	
	var challenger = state.get_player_by_id(challenger_id)
	var actor = state.get_player_by_id(state.last_move.actor_id)
	
	if not challenger or not actor:
		return {"success": false, "error": "Player not found"}
	
	var liar_was_lying = actor.current_hero != state.last_move.declared_alibi
	if liar_was_lying:
		# revert gains
		actor.coins -= state.last_move.coins_delta
		actor.gems -= state.last_move.gems_delta
		# restore monster if needed
		if state.last_move.type == "fight" and state.last_move.monster_id != "":
			state.active_monsters.append(state.last_move.monster_id)
		actor.life -= 1
		actor.current_hero = ""
	else:
		challenger.life -= 1
	
	# Clear last move
	state.last_move = null
	
	# Check for game end conditions
	_detect_outcome(state)
	
	return {
		"success": true,
		"challenge_correct": liar_was_lying,
		"actor_life": actor.life,
		"challenger_life": challenger.life
	}

# Check for game end conditions
func check_game_end(state: GameModels.GameState) -> Dictionary:
	_detect_outcome(state)
	
	if state.winner_ids.size() > 0:
		return {
			"game_ended": true,
			"winner_ids": state.winner_ids,
			"reason": "victory_condition_met"
		}
	
	if state.draw:
		return {
			"game_ended": true,
			"draw": true,
			"reason": "stalemate"
		}
	
	return {"game_ended": false}

# Get player-specific view of game state
func get_player_view(state: GameModels.GameState, player_id: String) -> Dictionary:
	var player = state.get_player_by_id(player_id)
	if not player:
		return {}
	
	var view = state.to_dict()
	
	# Hide other players' hands
	for p in view.players:
		if p.id != player_id:
			p.hand = []  # Hide hand
	
	# Hide burned cards
	view.burned_cards = []
	
	return view

# Get card data by ID
func get_card_data(card_id: String) -> Dictionary:
	return cards_data.get(card_id, {})

# Get power data by ID
func get_power_data(power_id: String) -> Dictionary:
	return powers_data.get(power_id, {})

# Get all cards
func get_all_cards() -> Array:
	return cards_data.values()

# Get all powers
func get_all_powers() -> Array:
	return powers_data.values()

# Get cards by type
func get_cards_by_type(card_type: String) -> Array:
	var filtered_cards = []
	for card in cards_data.values():
		if card.type == card_type:
			filtered_cards.append(card)
	return filtered_cards

# Get hero cards
func get_hero_cards() -> Array:
	return get_cards_by_type("hero")

# Get monster cards
func get_monster_cards() -> Array:
	return get_cards_by_type("monster")

# Forfeit functionality (from Go backend)
func forfeit_game(state: GameModels.GameState, player_id: String) -> Dictionary:
	var player = state.get_player_by_id(player_id)
	if not player:
		return {"success": false, "error": "Player not found"}
	
	if player.life == 0:
		return {"success": false, "error": "Already out"}
	
	# Mark dead, move to spectators
	player.life = 0
	state.spectator_ids.append(player_id)
	
	# Check for game end conditions
	_detect_outcome(state)
	
	return {
		"success": true,
		"message": "Player forfeited",
		"player_id": player_id
	}

# Initialize game with proper life values based on player count (from Go backend)
func initialize_game_with_life(state: GameModels.GameState) -> Dictionary:
	var player_count = state.players.size()
	var life_value = _get_life_for_players(player_count)
	
	for player in state.players:
		player.life = life_value
	
	# Deal initial hero cards
	for player in state.players:
		if state.hero_deck.size() > 0:
			player.current_hero = state.hero_deck[0]
			state.hero_deck.remove_at(0)
	
	# Draw 2 active monsters
	if state.monster_deck.size() < 2:
		return {"success": false, "error": "Not enough monster cards"}
	
	state.active_monsters = [state.monster_deck[0], state.monster_deck[1]]
	state.monster_deck = state.monster_deck.slice(2)
	
	# Randomize play order
	state.turn_order.shuffle()
	state.current_turn = state.turn_order[0]
	
	return {"success": true, "message": "Game initialized"}

func _get_life_for_players(n: int) -> int:
	if n <= 5:
		return 2
	elif n <= 10:
		return 3
	else:
		return 4

# Power effect implementations (from Go backend)
func _change_monster_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if state.monster_deck.size() == 0 or state.active_monsters.size() == 0:
		return {"success": false, "error": "No monsters to swap"}
	
	var idx = randi() % state.active_monsters.size()
	var draw = state.monster_deck[0]
	state.monster_deck.remove_at(0)
	state.active_monsters[idx] = draw
	
	return {"success": true}

func _fight_monster_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	return {"success": true}

func _add_strength_to_challenger_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	actor.bonus_strength = 1
	return {"success": true}

func _gain_2gems_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	actor.gems += 2
	
	# Trigger gems_gained event
	var event = Event.new("gems_gained", actor.id, "", 2)
	dispatch_event(state, event)
	
	return {"success": true}

func _steal_gem_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target or target.gems == 0:
		return {"success": false, "error": "Target empty"}
	
	# Check for steal_attempt event
	var event = Event.new("steal_attempt", actor.id, target.id, 1)
	if dispatch_event(state, event):
		return {"success": false, "error": "Steal blocked"}
	
	# Success
	target.gems -= 1
	actor.gems += 1
	
	# Trigger gems_gained event
	var gems_event = Event.new("gems_gained", actor.id, "", 1)
	dispatch_event(state, gems_event)
	
	return {"success": true}

# Additional power effects (from Go backend)
func _shuffle_heroes_deck_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	state.hero_deck.append_array(state.discard_pile)
	state.discard_pile.clear()
	state.hero_deck.shuffle()
	return {"success": true}

func _change_hero_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if not _ensure_hero_deck_not_empty(state):
		return {"success": false, "error": "Hero deck empty and no cards to recycle"}
	
	if actor.current_hero != "":
		state.discard_pile.append(actor.current_hero)
		state.public_discard = actor.current_hero
	
	actor.current_hero = state.hero_deck[0]
	state.hero_deck.remove_at(0)
	
	return {"success": true}

func _exchange_gems_with_player_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	var temp = actor.gems
	actor.gems = target.gems
	target.gems = temp
	
	return {"success": true}

func _exchange_2coins_for_life_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if actor.coins < 2:
		return {"success": false, "error": "Need 2 coins"}
	
	actor.coins -= 2
	actor.life += 1
	
	return {"success": true}

func _gain_life_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	actor.life += 1
	return {"success": true}

func _see_player_card_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	return {"success": true, "target_hero": target.current_hero}

func _remove_coin_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target or target.coins == 0:
		return {"success": false, "error": "Target has no coins"}
	
	target.coins -= 1
	return {"success": true}

func _shoot_player_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	var ok = (move.guess_hero != "" and move.guess_hero == target.current_hero)
	if ok:
		target.life -= 1
	
	state.last_shoot_ok = ok
	return {"success": true, "hit": ok}

func _all_in_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if not state.last_shoot_ok:
		actor.gems = 0
	else:
		actor.gems *= 2
	
	return {"success": true}

func _parry_this_you_f_casual_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	target.life -= 1
	return {"success": true}

func _steal_coin_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target or target.coins == 0:
		return {"success": false, "error": "Target empty"}
	
	target.coins -= 1
	actor.coins += 1
	
	return {"success": true}

func _mimic_power_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	var card = cards_data.get(target.current_hero)
	if not card:
		return {"success": false, "error": "Target hero not found"}
	
	var power_ids = card.get("power_ids", [])
	if power_ids.size() == 0:
		return {"success": false, "error": "Target hero has no powers"}
	
	var power = powers_data.get(power_ids[0])  # order-0
	if not power:
		return {"success": false, "error": "Power not found"}
	
	if power.get("type") == "passive":
		return {"success": false, "error": "Target power is passive"}
	
	# Execute the mimicked power
	return execute_power(state, actor, power, move)

func _dual_attack_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	state.dual_attack = true
	return {"success": true}

func _blind_draw_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if not _ensure_hero_deck_not_empty(state):
		return {"success": false, "error": "Hero deck empty and no cards to recycle"}
	
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	state.discard_pile.append(target.current_hero)
	target.current_hero = state.hero_deck[0]
	state.hero_deck.remove_at(0)
	target.current_alibi = ""  # must bluff again
	
	return {"success": true}

func _force_transform_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if not _ensure_hero_deck_not_empty(state):
		return {"success": false, "error": "Hero deck empty and no cards to recycle"}
	
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	state.discard_pile.append(target.current_hero)
	target.current_hero = state.hero_deck[0]
	state.hero_deck.remove_at(0)
	target.current_alibi = ""
	
	return {"success": true}

func _execution_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	if move.discard_card_id == "":
		return {"success": false, "error": "discard_card_id required for execution"}
	
	var card = cards_data.get(move.discard_card_id)
	if not card:
		return {"success": false, "error": "Card not found"}
	
	var is_unique = card.get("default_amount_per_session", 1) == 1
	var players_with_card = _find_players_with_card(state, move.discard_card_id)
	
	# Execute the card based on uniqueness
	var card_executed = false
	
	if is_unique and players_with_card.size() > 0:
		card_executed = _execute_unique_card(state, move.discard_card_id, players_with_card)
	else:
		card_executed = _execute_non_unique_card(state, move.discard_card_id, players_with_card)
	
	# Clear public discard if it was the executed card
	if state.public_discard == move.discard_card_id:
		state.public_discard = ""
	
	# Add the card to burned pile only if it was actually executed
	if card_executed:
		state.burned_cards.append(move.discard_card_id)
	
	return {"success": true, "card_executed": card_executed}

func _fight_player_with_discarded_card_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload) -> Dictionary:
	var target = state.get_player_by_id(move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	if move.discard_card_id == "":
		return {"success": false, "error": "discard_card_id missing"}
	
	var card = cards_data.get(move.discard_card_id)
	if not card:
		return {"success": false, "error": "Card not found"}
	
	var att = card.get("strength", 0) + actor.bonus_strength
	actor.bonus_strength = 0
	
	var def = _effective_strength(state, target)
	
	if att > def:
		target.life -= 1
	else:
		actor.life -= 1
	
	return {"success": true}

# Helper functions for execution
func _find_players_with_card(state: GameModels.GameState, card_id: String) -> Array[GameModels.PlayerState]:
	var players: Array[GameModels.PlayerState] = []
	for player in state.players:
		if player.current_hero == card_id:
			players.append(player)
	return players

func _execute_card_from_player(state: GameModels.GameState, player: GameModels.PlayerState) -> bool:
	if not _ensure_hero_deck_not_empty(state):
		return false
	
	state.discard_pile.append(player.current_hero)
	player.current_hero = state.hero_deck[0]
	state.hero_deck.remove_at(0)
	player.current_alibi = ""  # must bluff again
	player.life -= 1  # lose a life point
	return true

func _execute_unique_card(state: GameModels.GameState, card_id: String, players_with_card: Array[GameModels.PlayerState]) -> bool:
	if players_with_card.size() == 0:
		return false
	
	# Randomly choose one player if multiple have the same card
	var chosen_player = players_with_card[0]
	if players_with_card.size() > 1:
		chosen_player = players_with_card[randi() % players_with_card.size()]
	
	return _execute_card_from_player(state, chosen_player)

func _execute_non_unique_card(state: GameModels.GameState, card_id: String, players_with_card: Array[GameModels.PlayerState]) -> bool:
	# Try discard pile first
	var index = state.discard_pile.find(card_id)
	if index != -1:
		state.discard_pile.remove_at(index)
		return true
	
	# Try deck next
	index = state.hero_deck.find(card_id)
	if index != -1:
		state.hero_deck.remove_at(index)
		return true
	
	# Last resort: target a random player with the card
	if players_with_card.size() > 0:
		var chosen_player = players_with_card[randi() % players_with_card.size()]
		return _execute_card_from_player(state, chosen_player)
	
	return false

func _effective_strength(state: GameModels.GameState, player: GameModels.PlayerState) -> int:
	var card = cards_data.get(player.current_alibi)
	if not card:
		return 0
	
	var base = card.get("strength", 0)
	if player.bonus_strength != 0:
		base = player.bonus_strength
	return base

# Trigger handlers (from Go backend)
func _keep_gems_trigger(state: GameModels.GameState, event: Event):
	var target = state.get_player_by_id(event.target_id)
	if target and _has_passive(state, target, "keep_gems"):
		event.cancel = true

func _tribute_trigger(state: GameModels.GameState, event: Event):
	# Find the Mad Prince (player with tribute passive)
	var prince: GameModels.PlayerState = null
	for player in state.players:
		if _has_passive(state, player, "tribute"):
			prince = player
			break
	
	# If prince is alive and gems gained > 1, halve them and send to prince
	if prince and prince.life > 0 and event.amount > 1:
		var halved_amount = event.amount / 2
		prince.gems += halved_amount
		# The original player still gets the full amount (tribute is additional, not replacement)

func _teamwork_trigger(state: GameModels.GameState, event: Event):
	for player in state.players:
		if player.id == event.actor_id:
			continue
		if _has_passive(state, player, "teamwork"):
			player.gems += event.amount 