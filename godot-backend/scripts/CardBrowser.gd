extends Control

# Card Browser - Displays all cards organized by category and sorted by strength
# This scene provides a comprehensive view of all available cards in the game

signal card_selected(card_data: Dictionary)

# UI References - will be set in _ready()
var refresh_button: Button
var loading_label: Label
var tab_container: TabContainer

# All Cards Tab
var heroes_grid: GridContainer
var monsters_grid: GridContainer
var heroes_count: Label
var monsters_count: Label

# Heroes Only Tab
var heroes_only_grid: GridContainer
var heroes_only_count: Label

# Monsters Only Tab
var monsters_only_grid: GridContainer
var monsters_only_count: Label

# Card Detail Panel
var card_detail_panel: Panel
var card_image: TextureRect
var card_name: Label
var card_type: Label
var card_strength: Label
var card_description: RichTextLabel
var powers_list: VBoxContainer

# Card data storage
var all_cards: Array = []
var hero_cards: Array = []
var monster_cards: Array = []

# Card widget scene for reuse
var card_widget_scene: PackedScene

# Game logic instance
var game_logic: GameLogic

func _ready():
	# Set up node references
	_setup_node_references()
	
	# Wait for config and database to be ready
	await ConfigManager.config_loaded
	await Database.connection_established
	
	# Create game logic instance
	game_logic = GameLogic.new()
	
	# Load card widget scene
	card_widget_scene = preload("res://scenes/CardWidget.tscn")
	
	# Connect signals
	if refresh_button:
		refresh_button.pressed.connect(_on_refresh_button_pressed)
	
	# Load cards on startup
	await load_cards()

func _setup_node_references():
	# UI References
	refresh_button = get_node_or_null("VBoxContainer/Header/RefreshButton")
	loading_label = get_node_or_null("VBoxContainer/Header/LoadingLabel")
	tab_container = get_node_or_null("VBoxContainer/TabContainer")

	# All Cards Tab
	heroes_grid = get_node_or_null("VBoxContainer/TabContainer/All Cards/AllCardsVBox/HeroesSection/HeroesGrid")
	monsters_grid = get_node_or_null("VBoxContainer/TabContainer/All Cards/AllCardsVBox/MonstersSection/MonstersGrid")
	heroes_count = get_node_or_null("VBoxContainer/TabContainer/All Cards/AllCardsVBox/HeroesSection/HeroesHeader/HeroesCount")
	monsters_count = get_node_or_null("VBoxContainer/TabContainer/All Cards/AllCardsVBox/MonstersSection/MonstersHeader/MonstersCount")

	# Heroes Only Tab
	heroes_only_grid = get_node_or_null("VBoxContainer/TabContainer/Heroes Only/HeroesOnlyVBox/HeroesOnlyGrid")
	heroes_only_count = get_node_or_null("VBoxContainer/TabContainer/Heroes Only/HeroesOnlyVBox/HeroesOnlyHeader/HeroesOnlyCount")

	# Monsters Only Tab
	monsters_only_grid = get_node_or_null("VBoxContainer/TabContainer/Monsters Only/MonstersOnlyVBox/MonstersOnlyGrid")
	monsters_only_count = get_node_or_null("VBoxContainer/TabContainer/Monsters Only/MonstersOnlyVBox/MonstersOnlyHeader/MonstersOnlyCount")

	# Card Detail Panel
	card_detail_panel = get_node_or_null("CardDetailPanel")
	card_image = get_node_or_null("CardDetailPanel/CardDetailVBox/DetailContent/CardImage")
	card_name = get_node_or_null("CardDetailPanel/CardDetailVBox/DetailContent/CardName")
	card_type = get_node_or_null("CardDetailPanel/CardDetailVBox/DetailContent/CardType")
	card_strength = get_node_or_null("CardDetailPanel/CardDetailVBox/DetailContent/CardStrength")
	card_description = get_node_or_null("CardDetailPanel/CardDetailVBox/DetailContent/CardDescription")
	powers_list = get_node_or_null("CardDetailPanel/CardDetailVBox/DetailContent/PowersSection/PowersList")

func _on_refresh_button_pressed():
	await load_cards()

func load_cards():
	if loading_label:
		loading_label.visible = true
	if refresh_button:
		refresh_button.disabled = true
	
	# Clear existing cards
	clear_all_grids()
	
	# Load cards from GameLogic
	all_cards = game_logic.get_all_cards()
	
	# Separate cards by type
	hero_cards.clear()
	monster_cards.clear()
	
	for card in all_cards:
		if card.type == "hero":
			hero_cards.append(card)
		elif card.type == "monster":
			monster_cards.append(card)
	
	# Sort cards by strength (strongest to weakest)
	hero_cards.sort_custom(func(a, b): return a.strength > b.strength)
	monster_cards.sort_custom(func(a, b): return a.strength > b.strength)
	
	# Populate grids
	if heroes_grid:
		populate_grid(heroes_grid, hero_cards)
	if monsters_grid:
		populate_grid(monsters_grid, monster_cards)
	if heroes_only_grid:
		populate_grid(heroes_only_grid, hero_cards)
	if monsters_only_grid:
		populate_grid(monsters_only_grid, monster_cards)
	
	# Update counts
	update_card_counts()
	
	if loading_label:
		loading_label.visible = false
	if refresh_button:
		refresh_button.disabled = false
	
	print("Loaded %d hero cards and %d monster cards" % [hero_cards.size(), monster_cards.size()])

func clear_all_grids():
	# Clear all grid containers
	if heroes_grid:
		for child in heroes_grid.get_children():
			child.queue_free()
	
	if monsters_grid:
		for child in monsters_grid.get_children():
			child.queue_free()
	
	if heroes_only_grid:
		for child in heroes_only_grid.get_children():
			child.queue_free()
	
	if monsters_only_grid:
		for child in monsters_only_grid.get_children():
			child.queue_free()

func populate_grid(grid: GridContainer, cards: Array):
	if not grid:
		return
		
	for card in cards:
		var card_widget = card_widget_scene.instantiate()
		grid.add_child(card_widget)
		card_widget.setup_card(card)
		card_widget.card_clicked.connect(_on_card_clicked)

func update_card_counts():
	if heroes_count:
		heroes_count.text = "(%d cards)" % hero_cards.size()
	if monsters_count:
		monsters_count.text = "(%d cards)" % monster_cards.size()
	if heroes_only_count:
		heroes_only_count.text = "(%d cards)" % hero_cards.size()
	if monsters_only_count:
		monsters_only_count.text = "(%d cards)" % monster_cards.size()

func _on_card_clicked(card_data: Dictionary):
	show_card_details(card_data)
	card_selected.emit(card_data)

func show_card_details(card_data: Dictionary):
	if not card_detail_panel:
		return
		
	# Update card detail panel
	if card_name:
		card_name.text = card_data.name
	if card_type:
		card_type.text = "Type: " + card_data.type.capitalize()
	if card_strength:
		card_strength.text = "Strength: " + str(card_data.strength)
	if card_description:
		card_description.text = card_data.description
	
	# Load card image if available
	if card_image:
		if card_data.has("cover_url") and card_data.cover_url != "":
			load_card_image(card_data.cover_url)
		else:
			# Use placeholder or default image
			card_image.texture = null
	
	# Load powers
	load_card_powers(card_data)
	
	# Show the panel
	card_detail_panel.visible = true

func load_card_image(image_url: String):
	if not card_image:
		return
		
	# Create HTTP request to load image from URL
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(_on_image_loaded)
	
	# Store the request reference so it doesn't get garbage collected
	card_image.set_meta("http_request", http_request)
	
	# Request the image
	var error = http_request.request(image_url)
	if error != OK:
		print("Failed to request image: " + image_url)
		# Fall back to placeholder
		_create_placeholder_image()

func _on_image_loaded(result: int, response_code: int, headers: PackedStringArray, body: PackedByteArray):
	if result != HTTPRequest.RESULT_SUCCESS or response_code != 200:
		print("Failed to load image: HTTP " + str(response_code))
		_create_placeholder_image()
		return
	
	# Create image from the downloaded data
	var image = Image.new()
	var error = image.load_png_from_buffer(body)
	if error != OK:
		# Try loading as JPEG
		error = image.load_jpg_from_buffer(body)
		if error != OK:
			print("Failed to decode image data")
			_create_placeholder_image()
			return
	
	# Create texture from image
	var texture = ImageTexture.create_from_image(image)
	if card_image:
		card_image.texture = texture
	
	# Clean up the HTTP request
	if card_image and card_image.has_meta("http_request"):
		var http_request = card_image.get_meta("http_request")
		http_request.queue_free()
		card_image.remove_meta("http_request")

func _create_placeholder_image():
	if not card_image:
		return
		
	# Create a colored rectangle as placeholder
	var image = Image.create(200, 300, false, Image.FORMAT_RGBA8)
	image.fill(Color(0.3, 0.3, 0.4, 1.0))
	
	# Add some text to indicate it's a placeholder
	var font = ThemeDB.fallback_font
	var font_size = ThemeDB.fallback_font_size
	if font:
		# Draw placeholder text
		font.draw_string(image, Vector2(10, 150), "No Image", HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, Color.WHITE)
	
	card_image.texture = ImageTexture.create_from_image(image)

func load_card_powers(card_data: Dictionary):
	if not powers_list:
		return
		
	# Clear existing powers
	for child in powers_list.get_children():
		child.queue_free()
	
	# Load powers for this card
	if card_data.has("power_ids") and card_data.power_ids.size() > 0:
		for power_id in card_data.power_ids:
			var power_data = game_logic.get_power_data(power_id)
			if not power_data.is_empty():
				var power_label = Label.new()
				power_label.text = "• " + power_data.name + " (" + power_data.type + ")"
				power_label.add_theme_color_override("font_color", Color(0.8, 0.8, 1.0))
				powers_list.add_child(power_label)
				
				var desc_label = Label.new()
				desc_label.text = "  " + power_data.description
				desc_label.add_theme_color_override("font_color", Color(0.7, 0.7, 0.8))
				desc_label.autowrap_mode = TextServer.AUTOWRAP_WORD_SMART
				powers_list.add_child(desc_label)
	else:
		var no_powers_label = Label.new()
		no_powers_label.text = "No powers"
		no_powers_label.add_theme_color_override("font_color", Color(0.6, 0.6, 0.6))
		powers_list.add_child(no_powers_label)

func _on_close_button_pressed():
	if card_detail_panel:
		card_detail_panel.visible = false

# Public methods for external access
func get_hero_cards() -> Array:
	return hero_cards

func get_monster_cards() -> Array:
	return monster_cards

func get_all_cards() -> Array:
	return all_cards

func get_cards_by_strength_range(min_strength: int, max_strength: int) -> Array:
	var filtered_cards: Array = []
	for card in all_cards:
		if card.strength >= min_strength and card.strength <= max_strength:
			filtered_cards.append(card)
	return filtered_cards

func get_cards_by_type(card_type: String) -> Array:
	if card_type == "hero":
		return hero_cards
	elif card_type == "monster":
		return monster_cards
	else:
		return []

func search_cards(search_term: String) -> Array:
	var results: Array = []
	search_term = search_term.to_lower()
	
	for card in all_cards:
		if card.name.to_lower().contains(search_term) or \
		   card.description.to_lower().contains(search_term):
			results.append(card)
	
	return results 
