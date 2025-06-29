extends RefCounted

# Power Engine - Implements the complete power system from the Go backend
# This includes all active and passive powers, trigger system, and power execution logic

signal power_executed(power_id: String, actor_id: String, target_id: String, success: bool)

# Power registry - maps action names to effect functions
var power_registry: Dictionary = {}

# Trigger registry - maps event types to handler functions
var trigger_registry: Dictionary = {}

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

# Initialize the power system
func _init():
	_register_all_powers()
	_register_all_triggers()

# Register all active powers
func _register_all_powers():
	# change_monster - swap one active monster with top of deck
	_register_power("change_monster", _change_monster_effect)
	
	# fight_monster - handled elsewhere by "fight" move, power just flags free fight
	_register_power("fight_monster", _fight_monster_effect)
	
	# fight_player_with_discarded_card - strength duel with discarded hero
	_register_power("fight_player_with_discarded_card", _fight_player_with_discarded_card_effect)
	
	# add_strength_to_challenger - +1 only for same-turn fight_player_with_discarded_card
	_register_power("add_strength_to_challenger", _add_strength_to_challenger_effect)
	
	# shuffle_heroes_deck - put all non-burned discards back & shuffle
	_register_power("shuffle_heroes_deck", _shuffle_heroes_deck_effect)
	
	# change_hero - draw new hero, discard current
	_register_power("change_hero", _change_hero_effect)
	
	# exchange_gems_with_player
	_register_power("exchange_gems_with_player", _exchange_gems_with_player_effect)
	
	# exchange_2coins_for_life
	_register_power("exchange_2coins_for_life", _exchange_2coins_for_life_effect)
	
	# gain_2gems
	_register_power("gain_2gems", _gain_2gems_effect)
	
	# gain_life
	_register_power("gain_life", _gain_life_effect)
	
	# see_player_card
	_register_power("see_player_card", _see_player_card_effect)
	
	# remove_coin
	_register_power("remove_coin", _remove_coin_effect)
	
	# shoot_player
	_register_power("shoot_player", _shoot_player_effect)
	
	# all_in
	_register_power("all_in", _all_in_effect)
	
	# steal_gem (order 0 and 1)
	_register_power("steal_gem_0", _steal_gem_effect)
	_register_power("steal_gem_1", _steal_gem_effect)
	
	# parry_this_you_f_casual
	_register_power("parry_this_you_f_casual", _parry_this_you_f_casual_effect)
	
	# steal_coin
	_register_power("steal_coin", _steal_coin_effect)
	
	# mimic_power
	_register_power("mimic_power", _mimic_power_effect)
	
	# dual_attack
	_register_power("dual_attack", _dual_attack_effect)
	
	# blind_draw
	_register_power("blind_draw", _blind_draw_effect)
	
	# force_transform
	_register_power("force_transform", _force_transform_effect)
	
	# execution - burn a hero card permanently from the session
	_register_power("execution", _execution_effect)

# Register all trigger handlers
func _register_all_triggers():
	# keep_gems passive → cancel steal_attempt
	_register_trigger("steal_attempt", _keep_gems_trigger)
	
	# tribute passive → halve gems gained by other players and send to prince
	_register_trigger("gems_gained", _tribute_trigger)
	
	# teamwork passive → duplicate gems on gems_gained
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
func execute_power(state: GameModels.GameState, actor: GameModels.PlayerState, power: Dictionary, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var action = power.get("action", "")
	
	if not power_registry.has(action):
		return {"success": false, "error": "Unknown power action: " + action}
	
	var effect_func = power_registry[action]
	var result = effect_func.call(state, actor, move, cards_data, powers_data)
	
	power_executed.emit(power.get("id", ""), actor.id, move.target_player, result.get("success", false))
	
	return result

# Dispatch an event to trigger handlers
func dispatch_event(state: GameModels.GameState, event: Event, cards_data: Dictionary, powers_data: Dictionary) -> bool:
	if not trigger_registry.has(event.type):
		return false
	
	var handlers = trigger_registry[event.type]
	for handler in handlers:
		handler.call(state, event, cards_data, powers_data)
		if event.cancel:
			return true
	
	return false

# Helper functions
func _find_player(state: GameModels.GameState, player_id: String) -> GameModels.PlayerState:
	return state.get_player_by_id(player_id)

func _ensure_hero_deck_not_empty(state: GameModels.GameState) -> bool:
	if state.hero_deck.size() == 0:
		if state.discard_pile.size() > 0:
			state.hero_deck.append_array(state.discard_pile)
			state.discard_pile.clear()
			state.public_discard = ""  # clear public discard since we're recycling
			state.hero_deck.shuffle()
			return true
		else:
			return false  # no cards to recycle
	return true

func _has_passive(state: GameModels.GameState, player: GameModels.PlayerState, action: String, cards_data: Dictionary, powers_data: Dictionary) -> bool:
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

# Power effect implementations
func _change_monster_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	if state.monster_deck.size() == 0 or state.active_monsters.size() == 0:
		return {"success": false, "error": "No monsters to swap"}
	
	var idx = randi() % state.active_monsters.size()
	var draw = state.monster_deck[0]
	state.monster_deck.remove_at(0)
	state.active_monsters[idx] = draw
	
	return {"success": true}

func _fight_monster_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	# ApplyMove already does a fight; cost 1 gem taken by caller
	return {"success": true}

func _fight_player_with_discarded_card_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	if move.discard_card_id == "":
		return {"success": false, "error": "discard_card_id missing"}
	
	var card = cards_data.get(move.discard_card_id)
	if not card:
		return {"success": false, "error": "Card not found"}
	
	var att = card.get("strength", 0) + actor.bonus_strength
	actor.bonus_strength = 0
	
	var def = _effective_strength(state, target, cards_data)
	
	if att > def:
		target.life -= 1
	else:
		actor.life -= 1
	
	return {"success": true}

func _add_strength_to_challenger_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	actor.bonus_strength = 1
	return {"success": true}

func _shuffle_heroes_deck_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	state.hero_deck.append_array(state.discard_pile)
	state.discard_pile.clear()
	state.hero_deck.shuffle()
	return {"success": true}

func _change_hero_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	if not _ensure_hero_deck_not_empty(state):
		return {"success": false, "error": "Hero deck empty and no cards to recycle"}
	
	if actor.current_hero != "":
		state.discard_pile.append(actor.current_hero)
		state.public_discard = actor.current_hero
	
	actor.current_hero = state.hero_deck[0]
	state.hero_deck.remove_at(0)
	
	return {"success": true}

func _exchange_gems_with_player_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	var temp = actor.gems
	actor.gems = target.gems
	target.gems = temp
	
	return {"success": true}

func _exchange_2coins_for_life_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	if actor.coins < 2:
		return {"success": false, "error": "Need 2 coins"}
	
	actor.coins -= 2
	actor.life += 1
	
	return {"success": true}

func _gain_2gems_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	actor.gems += 2
	
	# Trigger gems_gained event
	var event = Event.new("gems_gained", actor.id, "", 2)
	dispatch_event(state, event, cards_data, powers_data)
	
	return {"success": true}

func _gain_life_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	actor.life += 1
	return {"success": true}

func _see_player_card_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	# In real implementation send secret notification; here: no state change
	# The target's current_hero is visible to the actor
	return {"success": true, "target_hero": target.current_hero}

func _remove_coin_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target or target.coins == 0:
		return {"success": false, "error": "Target has no coins"}
	
	target.coins -= 1
	return {"success": true}

func _shoot_player_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	var ok = (move.guess_hero != "" and move.guess_hero == target.current_hero)
	if ok:
		target.life -= 1
	
	state.last_shoot_ok = ok
	return {"success": true, "hit": ok}

func _all_in_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	if not state.last_shoot_ok:
		actor.gems = 0
	else:
		actor.gems *= 2
	
	return {"success": true}

func _steal_gem_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target or target.gems == 0:
		return {"success": false, "error": "Target empty"}
	
	# Check for steal_attempt event
	var event = Event.new("steal_attempt", actor.id, target.id, 1)
	if dispatch_event(state, event, cards_data, powers_data):
		return {"success": false, "error": "Steal blocked"}
	
	# Success
	target.gems -= 1
	actor.gems += 1
	
	# Trigger gems_gained event
	var gems_event = Event.new("gems_gained", actor.id, "", 1)
	dispatch_event(state, gems_event, cards_data, powers_data)
	
	return {"success": true}

func _parry_this_you_f_casual_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	target.life -= 1
	return {"success": true}

func _steal_coin_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
	if not target or target.coins == 0:
		return {"success": false, "error": "Target empty"}
	
	target.coins -= 1
	actor.coins += 1
	
	return {"success": true}

func _mimic_power_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	var target = _find_player(state, move.target_player)
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
	return execute_power(state, actor, power, move, cards_data, powers_data)

func _dual_attack_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	state.dual_attack = true
	return {"success": true}

func _blind_draw_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	if not _ensure_hero_deck_not_empty(state):
		return {"success": false, "error": "Hero deck empty and no cards to recycle"}
	
	var target = _find_player(state, move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	state.discard_pile.append(target.current_hero)
	target.current_hero = state.hero_deck[0]
	state.hero_deck.remove_at(0)
	target.current_alibi = ""  # must bluff again
	
	return {"success": true}

func _force_transform_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
	if not _ensure_hero_deck_not_empty(state):
		return {"success": false, "error": "Hero deck empty and no cards to recycle"}
	
	var target = _find_player(state, move.target_player)
	if not target:
		return {"success": false, "error": "Target not found"}
	
	state.discard_pile.append(target.current_hero)
	target.current_hero = state.hero_deck[0]
	state.hero_deck.remove_at(0)
	target.current_alibi = ""
	
	return {"success": true}

func _execution_effect(state: GameModels.GameState, actor: GameModels.PlayerState, move: GameModels.MovePayload, cards_data: Dictionary, powers_data: Dictionary) -> Dictionary:
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

# Helper functions for execution
func _find_players_with_card(state: GameModels.GameState, card_id: String) -> Array[GameModels.PlayerState]:
	var players: Array[GameModels.PlayerState] = []
	for player in state.players:
		if player.current_hero == card_id:
			players.append(player)
	return players

func _remove_card_from_discard_pile(state: GameModels.GameState, card_id: String) -> bool:
	var index = state.discard_pile.find(card_id)
	if index != -1:
		state.discard_pile.remove_at(index)
		return true
	return false

func _remove_card_from_deck(state: GameModels.GameState, card_id: String) -> bool:
	var index = state.hero_deck.find(card_id)
	if index != -1:
		state.hero_deck.remove_at(index)
		return true
	return false

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
	if _remove_card_from_discard_pile(state, card_id):
		return true
	
	# Try deck next
	if _remove_card_from_deck(state, card_id):
		return true
	
	# Last resort: target a random player with the card
	if players_with_card.size() > 0:
		var chosen_player = players_with_card[randi() % players_with_card.size()]
		return _execute_card_from_player(state, chosen_player)
	
	return false

func _effective_strength(state: GameModels.GameState, player: GameModels.PlayerState, cards_data: Dictionary) -> int:
	var card = cards_data.get(player.current_alibi)
	if not card:
		return 0
	
	var base = card.get("strength", 0)
	if player.bonus_strength != 0:
		base = player.bonus_strength
	return base

# Trigger handlers
func _keep_gems_trigger(state: GameModels.GameState, event: Event, cards_data: Dictionary, powers_data: Dictionary):
	var target = _find_player(state, event.target_id)
	if target and _has_passive(state, target, "keep_gems", cards_data, powers_data):
		event.cancel = true

func _tribute_trigger(state: GameModels.GameState, event: Event, cards_data: Dictionary, powers_data: Dictionary):
	# Find the Mad Prince (player with tribute passive)
	var prince: GameModels.PlayerState = null
	for player in state.players:
		if _has_passive(state, player, "tribute", cards_data, powers_data):
			prince = player
			break
	
	# If prince is alive and gems gained > 1, halve them and send to prince
	if prince and prince.life > 0 and event.amount > 1:
		var halved_amount = event.amount / 2
		prince.gems += halved_amount
		# The original player still gets the full amount (tribute is additional, not replacement)

func _teamwork_trigger(state: GameModels.GameState, event: Event, cards_data: Dictionary, powers_data: Dictionary):
	for player in state.players:
		if player.id == event.actor_id:
			continue
		if _has_passive(state, player, "teamwork", cards_data, powers_data):
			player.gems += event.amount 