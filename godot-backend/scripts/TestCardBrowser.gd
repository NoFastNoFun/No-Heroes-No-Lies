extends Node

# Test script for CardBrowser functionality
# This script can be attached to any node to test the card browser

# Game logic instance
var game_logic: GameLogic

func _ready():
	# Create game logic instance
	game_logic = GameLogic.new()
	
	# Wait for game data to load
	await game_logic.data_loaded_signal
	
	# Test the card browser
	test_card_browser()

func test_card_browser():
	print("=== Testing Card Browser ===")
	
	# Get all cards
	var all_cards = game_logic.get_all_cards()
	print("Total cards loaded: ", all_cards.size())
	
	# Separate by type
	var hero_cards: Array[Dictionary] = []
	var monster_cards: Array[Dictionary] = []
	
	for card in all_cards:
		if card.type == "hero":
			hero_cards.append(card)
		elif card.type == "monster":
			monster_cards.append(card)
	
	print("Hero cards: ", hero_cards.size())
	print("Monster cards: ", monster_cards.size())
	
	# Sort by strength
	hero_cards.sort_custom(func(a, b): return a.strength > b.strength)
	monster_cards.sort_custom(func(a, b): return a.strength > b.strength)
	
	# Display top 5 strongest heroes
	print("\n=== Top 5 Strongest Heroes ===")
	for i in range(min(5, hero_cards.size())):
		var card = hero_cards[i]
		print("%d. %s (Strength: %d)" % [i + 1, card.name, card.strength])
	
	# Display top 5 strongest monsters
	print("\n=== Top 5 Strongest Monsters ===")
	for i in range(min(5, monster_cards.size())):
		var card = monster_cards[i]
		print("%d. %s (Strength: %d)" % [i + 1, card.name, card.strength])
	
	# Test power loading for first hero
	if hero_cards.size() > 0:
		var first_hero = hero_cards[0]
		print("\n=== Powers for %s ===" % first_hero.name)
		
		if first_hero.has("power_ids") and first_hero.power_ids.size() > 0:
			for power_id in first_hero.power_ids:
				var power_data = game_logic.get_power_data(power_id)
				if not power_data.is_empty():
					print("- %s (%s): %s" % [power_data.name, power_data.type, power_data.description])
				else:
					print("- Power ID %s not found" % power_id)
		else:
			print("No powers found")
	
	print("\n=== Card Browser Test Complete ===") 