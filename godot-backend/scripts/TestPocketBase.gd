extends Node

# Test script for PocketBase integration
# This can be used to verify that cards and powers are loading correctly

class_name TestPocketBase

signal test_completed(results: Dictionary)

# Game logic instance
var game_logic: GameLogic

func _ready():
	# Create game logic instance
	game_logic = GameLogic.new()
	
	# Wait for data to load
	await game_logic.data_loaded_signal

func test_pocketbase_integration():
	var results = {
		"database_connection": false,
		"cards_loaded": false,
		"powers_loaded": false,
		"card_count": 0,
		"power_count": 0,
		"errors": []
	}
	
	print("Starting PocketBase integration test...")
	
	# Test database connection
	if Database.is_connected:
		results.database_connection = true
		print("✅ Database connection successful")
	else:
		results.errors.append("Database connection failed")
		print("❌ Database connection failed")
		test_completed.emit(results)
		return
	
	# Test cards loading
	var cards = game_logic.get_all_cards()
	results.card_count = cards.size()
	
	if cards.size() > 0:
		results.cards_loaded = true
		print("✅ Cards loaded successfully: " + str(cards.size()) + " cards")
		
		# Show first few cards
		for i in range(min(3, cards.size())):
			var card = cards[i]
			print("  - " + card.name + " (" + card.type + ", Strength: " + str(card.strength) + ")")
	else:
		results.errors.append("No cards loaded")
		print("❌ No cards loaded")
	
	# Test powers loading
	var powers = game_logic.get_all_powers()
	results.power_count = powers.size()
	
	if powers.size() > 0:
		results.powers_loaded = true
		print("✅ Powers loaded successfully: " + str(powers.size()) + " powers")
		
		# Show first few powers
		for i in range(min(3, powers.size())):
			var power = powers[i]
			print("  - " + power.name + " (" + power.type + ", Cost: " + str(power.cost) + ")")
	else:
		results.errors.append("No powers loaded")
		print("❌ No powers loaded")
	
	# Test specific card lookup
	if cards.size() > 0:
		var first_card = cards[0]
		var card_data = game_logic.get_card_data(first_card.id)
		if not card_data.is_empty():
			print("✅ Card lookup working: " + card_data.name)
		else:
			results.errors.append("Card lookup failed")
			print("❌ Card lookup failed")
	
	# Test specific power lookup
	if powers.size() > 0:
		var first_power = powers[0]
		var power_data = game_logic.get_power_data(first_power.id)
		if not power_data.is_empty():
			print("✅ Power lookup working: " + power_data.name)
		else:
			results.errors.append("Power lookup failed")
			print("❌ Power lookup failed")
	
	# Test hero and monster filtering
	var hero_cards = game_logic.get_hero_cards()
	var monster_cards = game_logic.get_monster_cards()
	
	print("Hero cards: " + str(hero_cards.size()))
	print("Monster cards: " + str(monster_cards.size()))
	
	# Summary
	print("\n=== Test Summary ===")
	print("Database connection: " + ("✅" if results.database_connection else "❌"))
	print("Cards loaded: " + ("✅" if results.cards_loaded else "❌") + " (" + str(results.card_count) + ")")
	print("Powers loaded: " + ("✅" if results.powers_loaded else "❌") + " (" + str(results.power_count) + ")")
	
	if results.errors.size() > 0:
		print("Errors:")
		for error in results.errors:
			print("  - " + error)
	
	test_completed.emit(results)

func test_card_structure():
	print("\n=== Testing Card Structure ===")
	var cards = game_logic.get_all_cards()
	
	if cards.size() == 0:
		print("No cards to test")
		return
	
	var sample_card = cards[0]
	print("Sample card structure:")
	print("  ID: " + sample_card.id)
	print("  Name: " + sample_card.name)
	print("  Type: " + sample_card.type)
	print("  Description: " + sample_card.description)
	print("  Strength: " + str(sample_card.strength))
	print("  Power IDs: " + str(sample_card.power_ids))
	print("  Default Amount: " + str(sample_card.default_amount_per_session))
	print("  Cover URL: " + sample_card.cover_url)
	
	if sample_card.has("loot"):
		var loot = sample_card.loot
		print("  Loot - Coins: " + str(loot.coins) + ", Gems: " + str(loot.gems))

func test_power_structure():
	print("\n=== Testing Power Structure ===")
	var powers = game_logic.get_all_powers()
	
	if powers.size() == 0:
		print("No powers to test")
		return
	
	var sample_power = powers[0]
	print("Sample power structure:")
	print("  ID: " + sample_power.id)
	print("  Name: " + sample_power.name)
	print("  Type: " + sample_power.type)
	print("  Cost: " + str(sample_power.cost))
	print("  Action: " + sample_power.action)
	print("  Target: " + sample_power.target)
	print("  Trigger: " + sample_power.trigger)
	print("  Order: " + str(sample_power.order))
	print("  Description: " + sample_power.description)

func test_deck_creation():
	print("\n=== Testing Deck Creation ===")
	
	# Test with sample player IDs
	var player_ids = ["player1", "player2"]
	var game_state = await game_logic.initialize_game_session(player_ids)
	
	print("Game state initialized:")
	print("  Players: " + str(game_state.players.size()))
	print("  Hero deck size: " + str(game_state.hero_deck.size()))
	print("  Monster deck size: " + str(game_state.monster_deck.size()))
	
	# Show first few cards in hero deck
	if game_state.hero_deck.size() > 0:
		print("  First hero cards:")
		for i in range(min(3, game_state.hero_deck.size())):
			var card_id = game_state.hero_deck[i]
			var card = game_logic.get_card_data(card_id)
			if not card.is_empty():
				print("    - " + card.name + " (Strength: " + str(card.strength) + ")")
	
	# Show first few cards in monster deck
	if game_state.monster_deck.size() > 0:
		print("  First monster cards:")
		for i in range(min(3, game_state.monster_deck.size())):
			var card_id = game_state.monster_deck[i]
			var card = game_logic.get_card_data(card_id)
			if not card.is_empty():
				print("    - " + card.name + " (Strength: " + str(card.strength) + ")")
	
	# Show player hands
	for player in game_state.players:
		print("  Player " + player.id + " hand:")
		for card_id in player.hand:
			var card = game_logic.get_card_data(card_id)
			if not card.is_empty():
				print("    - " + card.name + " (Strength: " + str(card.strength) + ")") 