extends RefCounted

# Game models that replicate the Go data structures
# Replaces the Go models package

class_name GameModels

# Card mirrors the cards collection from PocketBase
class Card:
	var id: String
	var type: String  # "hero" or "monster"
	var name: String
	var description: String
	var strength: int
	var power_ids: Array[String]
	var loot: Loot
	var default_amount_per_session: int
	var cover_url: String
	
	func _init(data: Dictionary = {}):
		id = data.get("id", "")
		type = data.get("type", "")
		name = data.get("name", "")
		description = data.get("description", "")
		strength = data.get("strength", 0)
		power_ids = data.get("power_ids", [])
		loot = Loot.new(data.get("loot", {}))
		default_amount_per_session = data.get("default_amount_per_session", 1)
		cover_url = data.get("cover", "")
	
	func to_dict() -> Dictionary:
		return {
			"id": id,
			"type": type,
			"name": name,
			"description": description,
			"strength": strength,
			"power_ids": power_ids,
			"loot": loot.to_dict(),
			"default_amount_per_session": default_amount_per_session,
			"cover": cover_url
		}

# Power represents one active or passive ability.
# @Description Game power/ability definition
class Power:
	var id: String
	var name: String
	var type: String  # "active" | "passive"
	var cost: int
	var action: String
	var target: String
	var trigger: String
	var order: int  # execution order within a card
	var description: String
	
	func _init(data: Dictionary = {}):
		id = data.get("id", "")
		name = data.get("name", "")
		type = data.get("type", "")
		cost = data.get("cost", 0)
		action = data.get("action", "")
		target = data.get("target", "")
		trigger = data.get("trigger", "")
		order = data.get("order", 0)
		description = data.get("description", "")
	
	func to_dict() -> Dictionary:
		return {
			"id": id,
			"name": name,
			"type": type,
			"cost": cost,
			"action": action,
			"target": target,
			"trigger": trigger,
			"order": order,
			"description": description
		}

# Loot defines monster rewards.
# @Description Loot rewards from defeating monsters
class Loot:
	var coins: int
	var gems: int
	
	func _init(data: Dictionary = {}):
		coins = data.get("coins", 0)
		gems = data.get("gems", 0)
	
	func to_dict() -> Dictionary:
		return {
			"coins": coins,
			"gems": gems
		}

# PlayerState represents the state of a single player
class PlayerState:
	var id: String
	var life: int
	var coins: int
	var gems: int
	var glory: int
	var current_hero: String
	var current_alibi: String
	var hand: Array[String]
	var bonus_strength: int
	var alt_strength: bool
	
	func _init(data: Dictionary = {}):
		id = data.get("id", "")
		life = data.get("life", 0)
		coins = data.get("coins", 0)
		gems = data.get("gems", 0)
		glory = data.get("glory", 0)
		current_hero = data.get("current_hero", "")
		current_alibi = data.get("current_alibi", "")
		hand = data.get("hand", [])
		bonus_strength = data.get("bonus_strength", 0)
		alt_strength = data.get("alt_strength", false)
	
	func to_dict() -> Dictionary:
		return {
			"id": id,
			"life": life,
			"coins": coins,
			"gems": gems,
			"glory": glory,
			"current_hero": current_hero,
			"current_alibi": current_alibi,
			"hand": hand,
			"bonus_strength": bonus_strength,
			"alt_strength": alt_strength
		}

# LastMove is stored for 5 s, so others can challenge.
class LastMove:
	var actor_id: String
	var type: String  # "fight" or "power"
	var declared_alibi: String
	var monster_id: String
	var coins_delta: int
	var gems_delta: int
	var life_delta_id: String
	var life_delta: int
	var expires_at: int  # unix ms
	var correct: bool
	
	func _init(data: Dictionary = {}):
		actor_id = data.get("actor_id", "")
		type = data.get("type", "")
		declared_alibi = data.get("declared_alibi", "")
		monster_id = data.get("monster_id", "")
		coins_delta = data.get("coins_delta", 0)
		gems_delta = data.get("gems_delta", 0)
		life_delta_id = data.get("life_delta_id", "")
		life_delta = data.get("life_delta", 0)
		expires_at = data.get("expires_at", 0)
		correct = data.get("correct", false)
	
	func to_dict() -> Dictionary:
		return {
			"actor_id": actor_id,
			"type": type,
			"declared_alibi": declared_alibi,
			"monster_id": monster_id,
			"coins_delta": coins_delta,
			"gems_delta": gems_delta,
			"life_delta_id": life_delta_id,
			"life_delta": life_delta,
			"expires_at": expires_at,
			"correct": correct
		}
	
	func is_expired() -> bool:
		return Time.get_unix_time_from_system() * 1000 > expires_at

# GameState represents the complete game state
class GameState:
	var players: Array[PlayerState]
	var turn_order: Array[String]
	var current_turn: String
	var coin_pool: int
	var gem_pool: int
	var hero_deck: Array[String]
	var monster_deck: Array[String]
	var active_monsters: Array[String]
	var seed: int
	var last_move: LastMove
	var discard_pile: Array[String]
	var burned_cards: Array[String]
	var public_discard: String
	var order0_used_by: Dictionary
	var dual_attack: bool
	var last_shoot_ok: bool
	var winner_ids: Array[String]
	var draw: bool
	var pause_votes: Dictionary
	var ready_ids: Array[String]
	var spectator_ids: Array[String]
	
	func _init(data: Dictionary = {}):
		players = []
		if data.has("players"):
			for player_data in data.players:
				players.append(PlayerState.new(player_data))
		
		turn_order = data.get("turn_order", [])
		current_turn = data.get("current_turn", "")
		coin_pool = data.get("coin_pool", 0)
		gem_pool = data.get("gem_pool", 0)
		hero_deck = data.get("hero_deck", [])
		monster_deck = data.get("monster_deck", [])
		active_monsters = data.get("active_monsters", [])
		seed = data.get("seed", 0)
		
		if data.has("last_move") and data.last_move != null:
			last_move = LastMove.new(data.last_move)
		else:
			last_move = null
		
		discard_pile = data.get("discard_pile", [])
		burned_cards = data.get("burned_cards", [])
		public_discard = data.get("public_discard", "")
		order0_used_by = data.get("order0_used_by", {})
		dual_attack = data.get("dual_attack", false)
		last_shoot_ok = data.get("last_shoot_ok", false)
		winner_ids = data.get("winner_ids", [])
		draw = data.get("draw", false)
		pause_votes = data.get("pause_votes", {})
		ready_ids = data.get("ready_ids", [])
		spectator_ids = data.get("spectator_ids", [])
	
	func to_dict() -> Dictionary:
		var result = {
			"players": [],
			"turn_order": turn_order,
			"current_turn": current_turn,
			"coin_pool": coin_pool,
			"gem_pool": gem_pool,
			"hero_deck": hero_deck,
			"monster_deck": monster_deck,
			"active_monsters": active_monsters,
			"seed": seed,
			"discard_pile": discard_pile,
			"burned_cards": burned_cards,
			"public_discard": public_discard,
			"order0_used_by": order0_used_by,
			"dual_attack": dual_attack,
			"last_shoot_ok": last_shoot_ok,
			"winner_ids": winner_ids,
			"draw": draw,
			"pause_votes": pause_votes,
			"ready_ids": ready_ids,
			"spectator_ids": spectator_ids
		}
		
		for player in players:
			result.players.append(player.to_dict())
		
		if last_move != null:
			result.last_move = last_move.to_dict()
		else:
			result.last_move = null
		
		return result
	
	func get_player_by_id(player_id: String) -> PlayerState:
		for player in players:
			if player.id == player_id:
				return player
		return null
	
	func add_player(player: PlayerState):
		players.append(player)
		if not turn_order.has(player.id):
			turn_order.append(player.id)
	
	func remove_player(player_id: String):
		players = players.filter(func(p): return p.id != player_id)
		turn_order.erase(player_id)
		ready_ids.erase(player_id)
		spectator_ids.erase(player_id)
	
	func is_player_ready(player_id: String) -> bool:
		return ready_ids.has(player_id)
	
	func toggle_player_ready(player_id: String):
		if ready_ids.has(player_id):
			ready_ids.erase(player_id)
		else:
			ready_ids.append(player_id)
	
	func is_game_started() -> bool:
		return current_turn != ""
	
	func is_game_finished() -> bool:
		return winner_ids.size() > 0 or draw
	
	func get_current_player() -> PlayerState:
		if current_turn == "":
			return null
		return get_player_by_id(current_turn)
	
	func next_turn():
		if turn_order.size() == 0:
			return
		
		var current_index = turn_order.find(current_turn)
		if current_index == -1:
			current_turn = turn_order[0]
		else:
			var next_index = (current_index + 1) % turn_order.size()
			current_turn = turn_order[next_index]

# GameSession mirrors the PocketBase record we work with.
# @Description Game session information
class GameSession:
	var id: String
	var player_ids: Array[String]
	var state: GameState
	var is_active: bool
	var created_at: String
	var updated_at: String
	
	func _init(data: Dictionary = {}):
		id = data.get("id", "")
		player_ids = data.get("player_ids", [])
		state = GameState.new(data.get("state", {}))
		is_active = data.get("is_active", true)
		created_at = data.get("created_at", "")
		updated_at = data.get("updated_at", "")
	
	func to_dict() -> Dictionary:
		return {
			"id": id,
			"player_ids": player_ids,
			"state": state.to_dict(),
			"is_active": is_active,
			"created_at": created_at,
			"updated_at": updated_at
		}
	
	func add_player(player_id: String):
		if not player_ids.has(player_id):
			player_ids.append(player_id)
	
	func remove_player(player_id: String):
		player_ids.erase(player_id)
		state.remove_player(player_id)
	
	func is_player_in_session(player_id: String) -> bool:
		return player_ids.has(player_id) or state.spectator_ids.has(player_id)

# MovePayload represents a move submitted by a player
class MovePayload:
	var type: String  # "demask", "fight", "power"
	var target_player: String
	var guess: String
	var power_id: String
	var monster_id: String
	var declared_alibi: String
	var discard_card_id: String
	var guess_hero: String
	
	func _init(data: Dictionary = {}):
		type = data.get("type", "")
		target_player = data.get("target_player", "")
		guess = data.get("guess", "")
		power_id = data.get("power_id", "")
		monster_id = data.get("monster_id", "")
		declared_alibi = data.get("declared_alibi", "")
		discard_card_id = data.get("discard_card_id", "")
		guess_hero = data.get("guess_hero", "")
	
	func to_dict() -> Dictionary:
		return {
			"type": type,
			"target_player": target_player,
			"guess": guess,
			"power_id": power_id,
			"monster_id": monster_id,
			"declared_alibi": declared_alibi,
			"discard_card_id": discard_card_id,
			"guess_hero": guess_hero
		}
	
	func is_valid() -> bool:
		match type:
			"demask":
				return target_player != "" and guess_hero != ""
			"fight":
				return monster_id != "" and declared_alibi != ""
			"power":
				return power_id != ""
			_:
				return false 
