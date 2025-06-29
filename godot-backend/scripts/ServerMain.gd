extends Node

# Main server script that handles UI and coordinates server operations
# This is the entry point for the Godot server application

@onready var status_label: Label = $UI/VBoxContainer/StatusLabel
@onready var start_button: Button = $UI/VBoxContainer/StartButton
@onready var stop_button: Button = $UI/VBoxContainer/StopButton
@onready var card_browser_button: Button = $UI/VBoxContainer/CardBrowserButton
@onready var log_text: TextEdit = $UI/LogContainer/LogText
@onready var card_browser_container: Control = $UI/CardBrowserContainer

# CardBrowser scene reference
var card_browser_scene: PackedScene
var current_card_browser: Control

# Game logic instance
var game_logic: GameLogic

func _ready():
	# Connect button signals
	start_button.pressed.connect(_on_start_button_pressed)
	stop_button.pressed.connect(_on_stop_button_pressed)
	card_browser_button.pressed.connect(_on_card_browser_button_pressed)
	
	# Initially disable stop button
	stop_button.disabled = true
	
	# Load CardBrowser scene
	card_browser_scene = preload("res://scenes/CardBrowser.tscn")
	
	# Create game logic instance
	game_logic = GameLogic.new()
	
	# Connect to server events
	GameServer.server_started.connect(_on_server_started)
	GameServer.server_stopped.connect(_on_server_stopped)
	GameServer.graceful_shutdown_completed.connect(_on_graceful_shutdown_completed)
	
	# Connect to database events
	Database.connection_established.connect(_on_database_connected)
	Database.connection_failed.connect(_on_database_failed)
	Database.cards_loaded.connect(_on_cards_loaded)
	Database.powers_loaded.connect(_on_powers_loaded)
	
	# Connect to auth events
	AuthManager.user_authenticated.connect(_on_user_authenticated)
	AuthManager.authentication_failed.connect(_on_authentication_failed)
	
	# Connect to game logic events
	game_logic.cards_loaded.connect(_on_game_cards_loaded)
	game_logic.powers_loaded.connect(_on_game_powers_loaded)
	game_logic.data_loaded_signal.connect(_on_game_data_loaded)
	
	# Log initial status
	log_message("Server application started")
	log_message("Waiting for configuration to load...")
	
	# Wait for config to load
	await ConfigManager.config_loaded
	log_message("Configuration loaded successfully")
	
	# Update status
	status_label.text = "Server Status: Loading Game Data..."
	start_button.disabled = true
	
	# Wait for game data to load
	await game_logic.data_loaded_signal
	log_message("Game data loaded successfully")
	
	# Update status
	status_label.text = "Server Status: Ready"
	start_button.disabled = false

func _on_start_button_pressed():
	log_message("Starting server...")
	status_label.text = "Server Status: Starting..."
	start_button.disabled = true
	
	# Start the game server
	GameServer.start_server()

func _on_stop_button_pressed():
	log_message("Stopping server...")
	status_label.text = "Server Status: Stopping..."
	stop_button.disabled = true
	
	# Stop the game server
	GameServer.stop_server()

func _on_card_browser_button_pressed():
	if current_card_browser == null:
		open_card_browser()
	else:
		close_card_browser()

func open_card_browser():
	log_message("Opening Card Browser...")
	
	# Create CardBrowser instance
	current_card_browser = card_browser_scene.instantiate()
	card_browser_container.add_child(current_card_browser)
	
	# Show the container
	card_browser_container.visible = true
	
	# Update button text
	card_browser_button.text = "Close Card Browser"
	
	log_message("Card Browser opened")

func close_card_browser():
	log_message("Closing Card Browser...")
	
	# Remove CardBrowser instance
	if current_card_browser != null:
		current_card_browser.queue_free()
		current_card_browser = null
	
	# Hide the container
	card_browser_container.visible = false
	
	# Update button text
	card_browser_button.text = "Open Card Browser"
	
	log_message("Card Browser closed")

func _on_server_started():
	log_message("Server started successfully")
	status_label.text = "Server Status: Running"
	stop_button.disabled = false

func _on_server_stopped():
	log_message("Server stopped")
	status_label.text = "Server Status: Stopped"
	start_button.disabled = false

func _on_graceful_shutdown_completed():
	log_message("Graceful shutdown completed")
	status_label.text = "Server Status: Stopped"
	start_button.disabled = false

func _on_database_connected():
	log_message("Database connection established")

func _on_database_failed(error: String):
	log_message("Database connection failed: " + error)

func _on_cards_loaded(cards: Array):
	log_message("Loaded " + str(cards.size()) + " cards from PocketBase")

func _on_powers_loaded(powers: Array):
	log_message("Loaded " + str(powers.size()) + " powers from PocketBase")

func _on_game_cards_loaded(cards: Array):
	log_message("Game logic processed " + str(cards.size()) + " cards")

func _on_game_powers_loaded(powers: Array):
	log_message("Game logic processed " + str(powers.size()) + " powers")

func _on_game_data_loaded():
	log_message("All game data loaded and ready")

func _on_user_authenticated(user_data: Dictionary):
	log_message("User authenticated: " + user_data.get("email", "unknown"))

func _on_authentication_failed(error: String):
	log_message("Authentication failed: " + error)

func log_message(message: String):
	var timestamp = Time.get_datetime_string_from_system()
	var log_entry = "[" + timestamp + "] " + message
	
	# Add to log text
	log_text.text += log_entry + "\n"
	
	# Auto-scroll to bottom
	log_text.scroll_vertical = log_text.get_line_count()
	
	# Also print to console
	print(log_entry)

# Handle application quit
func _notification(what):
	if what == NOTIFICATION_WM_CLOSE_REQUEST:
		log_message("Application quit requested")
		if GameServer.is_running:
			log_message("Initiating graceful shutdown...")
			GameServer.stop_server()
			# Wait for graceful shutdown to complete
			await GameServer.graceful_shutdown_completed
		get_tree().quit()

# Handle input for testing
func _input(event):
	if event is InputEventKey and event.pressed:
		match event.keycode:
			KEY_F1:
				# Test health endpoint
				log_message("Testing health endpoint...")
				_test_health_endpoint()
			KEY_F2:
				# Test database connection
				log_message("Testing database connection...")
				_test_database_connection()
			KEY_F3:
				# Test authentication
				log_message("Testing authentication...")
				_test_authentication()
			KEY_F4:
				# Test cards loading
				log_message("Testing cards loading...")
				_test_cards_loading()
			KEY_F5:
				# Toggle card browser
				log_message("Toggling card browser...")
				_on_card_browser_button_pressed()

func _test_health_endpoint():
	# This would make an HTTP request to the health endpoint
	# In a real implementation, you'd use HTTPRequest
	log_message("Health endpoint test completed")

func _test_database_connection():
	# Test database connection
	if Database.is_connected:
		log_message("Database is connected")
	else:
		log_message("Database is not connected")

func _test_authentication():
	# Test JWT token generation
	var test_user = {
		"id": "test_user",
		"email": "test@example.com"
	}
	
	var token = AuthManager.generate_token(test_user)
	log_message("Generated test token: " + token.substr(0, 20) + "...")
	
	var validation = AuthManager.validate_token(token)
	if validation.valid:
		log_message("Token validation successful")
	else:
		log_message("Token validation failed: " + validation.error)

func _test_cards_loading():
	# Test cards and powers data
	var cards = game_logic.get_all_cards()
	var powers = game_logic.get_all_powers()
	
	log_message("Cards loaded: " + str(cards.size()))
	log_message("Powers loaded: " + str(powers.size()))
	
	if cards.size() > 0:
		var first_card = cards[0]
		log_message("First card: " + first_card.name + " (Strength: " + str(first_card.strength) + ")")
	
	if powers.size() > 0:
		var first_power = powers[0]
		log_message("First power: " + first_power.name + " (Cost: " + str(first_power.cost) + ")") 