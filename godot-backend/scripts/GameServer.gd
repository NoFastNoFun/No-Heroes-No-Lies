extends Node

# Main game server that handles HTTP requests
# Replaces the Go Chi router and server functionality

signal server_started
signal server_stopped
signal graceful_shutdown_completed

# HTTP server implementation using Godot's built-in HTTP server
var http_server: Node
var active_connections: Array[HTTPRequest] = []
var is_running: bool = false
var graceful_shutdown_timer: Timer

# Game state management
var game_sessions: Dictionary = {}
var session_timeout_timer: Timer

# Game logic instance
var game_logic: GameLogic

func _ready():
	# Wait for config to load
	await ConfigManager.config_loaded
	
	# Create game logic instance
	game_logic = GameLogic.new()
	
	# Set up graceful shutdown timer
	graceful_shutdown_timer = Timer.new()
	add_child(graceful_shutdown_timer)
	graceful_shutdown_timer.one_shot = true
	graceful_shutdown_timer.timeout.connect(_on_graceful_shutdown_timeout)
	
	# Set up session timeout timer
	session_timeout_timer = Timer.new()
	add_child(session_timeout_timer)
	session_timeout_timer.wait_time = 60.0  # Check every minute
	session_timeout_timer.timeout.connect(_cleanup_expired_sessions)
	session_timeout_timer.start()
	
	# Connect to database events
	Database.connection_established.connect(_on_database_connected)
	Database.connection_failed.connect(_on_database_failed)

func start_server():
	if is_running:
		print("Server is already running")
		return
	
	var port = ConfigManager.get_server_port()
	var host = ConfigManager.get_server_host()
	
	# Create HTTP server using Godot's built-in server
	http_server = _create_http_server()
	add_child(http_server)
	
	# Start listening
	var result = http_server.listen(port, host)
	if result != OK:
		print("Failed to start server on " + host + ":" + str(port))
		return
	
	# Set up request handling
	http_server.request_received.connect(_handle_request)
	
	is_running = true
	server_started.emit()
	print("Server started on " + host + ":" + str(port))

func _create_http_server() -> Node:
	# Create a simple HTTP server using Godot's built-in functionality
	var server = Node.new()
	server.set_script(preload("res://scripts/SimpleHTTPServer.gd"))
	return server

func stop_server():
	if not is_running:
		print("Server is not running")
		return
	
	print("Initiating graceful shutdown...")
	
	# Stop accepting new connections
	if http_server:
		http_server.stop()
		http_server.queue_free()
		http_server = null
	
	# Wait for active connections to complete
	graceful_shutdown_timer.start(ConfigManager.get_graceful_shutdown_timeout())
	
	# In a real implementation, you'd track active requests and wait for them to complete
	# For now, we'll just wait for the timeout

func _on_graceful_shutdown_timeout():
	is_running = false
	graceful_shutdown_completed.emit()
	print("Graceful shutdown completed")

func _handle_request(request: HTTPRequest):
	var method = request.get_method()
	var path = request.get_path()
	var headers = request.get_headers()
	var body = request.get_body()
	
	print("Received " + method + " request to " + path)
	
	# Handle CORS preflight
	if method == "OPTIONS":
		_handle_cors_preflight(request)
		return
	
	# Parse request body
	var request_data = {}
	if body.size() > 0:
		var json = JSON.new()
		var parse_result = json.parse(body.get_string_from_utf8())
		if parse_result == OK:
			request_data = json.data
	
	# Route the request
	var response = await _route_request(method, path, headers, request_data)
	
	# Send response
	_send_response(request, response)

func _route_request(method: String, path: String, headers: Dictionary, data: Dictionary) -> Dictionary:
	# Health check endpoints
	if path == "/api/health":
		return _handle_health_check()
	elif path == "/api/health/slow":
		return await _handle_slow_health_check()
	
	# Auth endpoints (no auth required)
	elif path == "/api/auth/login" and method == "POST":
		return _handle_login(data)
	elif path == "/api/auth/register" and method == "POST":
		return _handle_register(data)
	
	# Protected endpoints (require auth)
	else:
		var auth_header = headers.get("Authorization", "")
		var token = AuthManager.extract_token_from_header(auth_header)
		var auth_result = AuthManager.require_auth(token)
		
		if not auth_result.authorized:
			return {
				"status_code": 401,
				"headers": {"Content-Type": "application/json"},
				"body": JSON.stringify({"error": auth_result.error})
			}
		
		# Game session endpoints
		if path.begins_with("/api/game/"):
			return _handle_game_endpoints(method, path, data, auth_result.user_data)
		
		# Cards endpoints
		elif path.begins_with("/api/cards"):
			return _handle_cards_endpoints(method, path, data, auth_result.user_data)
		
		# Challenge endpoints
		elif path.begins_with("/api/challenge"):
			return _handle_challenge_endpoints(method, path, data, auth_result.user_data)
		
		# Unknown endpoint
		else:
			return {
				"status_code": 404,
				"headers": {"Content-Type": "application/json"},
				"body": JSON.stringify({"error": "Endpoint not found"})
			}

func _handle_health_check() -> Dictionary:
	return {
		"status_code": 200,
		"headers": {"Content-Type": "application/json"},
		"body": JSON.stringify({"status": "ok", "timestamp": Time.get_datetime_string_from_system()})
	}

func _handle_slow_health_check() -> Dictionary:
	# Simulate slow health check
	await get_tree().create_timer(2.0).timeout
	return {
		"status_code": 200,
		"headers": {"Content-Type": "application/json"},
		"body": JSON.stringify({"status": "ok", "slow": true, "timestamp": Time.get_datetime_string_from_system()})
	}

func _handle_login(data: Dictionary) -> Dictionary:
	var email = data.get("email", "")
	var password = data.get("password", "")
	
	if email == "" or password == "":
		return {
			"status_code": 400,
			"headers": {"Content-Type": "application/json"},
			"body": JSON.stringify({"error": "Email and password required"})
		}
	
	# In a real implementation, you'd validate against the database
	# For now, we'll simulate a successful login
	var user_data = {
		"id": "user_" + str(randi()),
		"email": email
	}
	
	var token = AuthManager.generate_token(user_data)
	var refresh_token = AuthManager.generate_refresh_token(user_data.id)
	
	return {
		"status_code": 200,
		"headers": {"Content-Type": "application/json"},
		"body": JSON.stringify({
			"token": token,
			"refresh_token": refresh_token,
			"user": user_data
		})
	}

func _handle_register(data: Dictionary) -> Dictionary:
	var email = data.get("email", "")
	var password = data.get("password", "")
	var password_confirm = data.get("password_confirm", "")
	
	if email == "" or password == "" or password_confirm == "":
		return {
			"status_code": 400,
			"headers": {"Content-Type": "application/json"},
			"body": JSON.stringify({"error": "Email, password, and password confirmation required"})
		}
	
	if password != password_confirm:
		return {
			"status_code": 400,
			"headers": {"Content-Type": "application/json"},
			"body": JSON.stringify({"error": "Passwords do not match"})
		}
	
	# In a real implementation, you'd create the user in the database
	# For now, we'll simulate a successful registration
	var user_data = {
		"id": "user_" + str(randi()),
		"email": email
	}
	
	var token = AuthManager.generate_token(user_data)
	var refresh_token = AuthManager.generate_refresh_token(user_data.id)
	
	return {
		"status_code": 201,
		"headers": {"Content-Type": "application/json"},
		"body": JSON.stringify({
			"token": token,
			"refresh_token": refresh_token,
			"user": user_data
		})
	}

func _handle_game_endpoints(method: String, path: String, data: Dictionary, user_data: Dictionary) -> Dictionary:
	var path_parts = path.split("/")
	
	if path_parts.size() == 3 and method == "POST":
		# POST /api/game - Create new game session
		return _create_game_session(user_data)
	
	elif path_parts.size() == 4 and path_parts[3] == "join" and method == "POST":
		# POST /api/game/{id}/join - Join game session
		var session_id = path_parts[2]
		return _join_game_session(session_id, user_data, data)
	
	elif path_parts.size() == 4 and path_parts[3] == "start" and method == "POST":
		# POST /api/game/{id}/start - Start game session
		var session_id = path_parts[2]
		return _start_game_session(session_id, user_data)
	
	elif path_parts.size() == 4 and path_parts[3] == "ready" and method == "POST":
		# POST /api/game/{id}/ready - Toggle player ready status
		var session_id = path_parts[2]
		return _toggle_player_ready(session_id, user_data)
	
	elif path_parts.size() == 3 and method == "GET":
		# GET /api/game/{id} - Get current game session state
		var session_id = path_parts[2]
		return _get_game_session(session_id, user_data)
	
	elif path_parts.size() == 4 and path_parts[3] == "move" and method == "POST":
		# POST /api/game/{id}/move - Submit a move
		var session_id = path_parts[2]
		return _submit_move(session_id, user_data, data)
	
	elif path_parts.size() == 4 and path_parts[3] == "forfeit" and method == "POST":
		# POST /api/game/{id}/forfeit - Forfeit the current game
		var session_id = path_parts[2]
		return _forfeit_game(session_id, user_data)
	
	else:
		return {
			"status_code": 404,
			"headers": {"Content-Type": "application/json"},
			"body": JSON.stringify({"error": "Game endpoint not found"})
		}

func _handle_cards_endpoints(method: String, path: String, data: Dictionary, user_data: Dictionary) -> Dictionary:
	if path == "/api/cards" and method == "GET":
		# GET /api/cards - Get all cards
		return _get_cards()
	
	elif path.begins_with("/api/cards/") and method == "GET":
		# GET /api/cards/{id} - Get specific card
		var card_id = path.split("/")[-1]
		return _get_card(card_id)
	
	else:
		return {
			"status_code": 404,
			"headers": {"Content-Type": "application/json"},
			"body": JSON.stringify({"error": "Cards endpoint not found"})
		}

func _handle_challenge_endpoints(method: String, path: String, data: Dictionary, user_data: Dictionary) -> Dictionary:
	if path.begins_with("/api/challenge/") and method == "POST":
		# POST /api/challenge/{id} - Resolve a challenge
		var challenge_id = path.split("/")[-1]
		return _resolve_challenge(challenge_id, user_data, data)
	
	else:
		return {
			"status_code": 404,
			"headers": {"Content-Type": "application/json"},
			"body": JSON.stringify({"error": "Challenge endpoint not found"})
		}

func _handle_cors_preflight(request: HTTPRequest):
	var response_headers = {
		"Access-Control-Allow-Origin": "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Accept, Authorization, Content-Type, X-CSRF-Token",
		"Access-Control-Max-Age": "300"
	}
	
	request.send_response(200, response_headers)

func _send_response(request: HTTPRequest, response: Dictionary):
	var status_code = response.get("status_code", 200)
	var headers = response.get("headers", {})
	var body = response.get("body", "")
	
	# Add CORS headers
	headers["Access-Control-Allow-Origin"] = "*"
	headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
	headers["Access-Control-Allow-Headers"] = "Accept, Authorization, Content-Type, X-CSRF-Token"
	
	request.send_response(status_code, headers)
	if body != "":
		request.send_body(body.to_utf8_buffer())

# Game session management functions
func _create_game_session(user_data: Dictionary) -> Dictionary:
	var session_id = "session_" + str(randi())
	var session = {
		"id": session_id,
		"player_ids": [user_data.id],
		"state": _create_initial_game_state(),
		"is_active": true,
		"created_at": Time.get_datetime_string_from_system(),
		"updated_at": Time.get_datetime_string_from_system()
	}
	
	game_sessions[session_id] = session
	
	return {
		"status_code": 201,
		"headers": {"Content-Type": "application/json"},
		"body": JSON.stringify(session)
	}

func _create_initial_game_state() -> Dictionary:
	return {
		"players": [],
		"turn_order": [],
		"current_turn": "",
		"coin_pool": 0,
		"gem_pool": 0,
		"hero_deck": [],
		"monster_deck": [],
		"active_monsters": [],
		"seed": randi(),
		"last_move": null,
		"discard_pile": [],
		"burned_cards": [],
		"public_discard": "",
		"order0_used_by": {},
		"dual_attack": false,
		"last_shoot_ok": false,
		"winner_ids": [],
		"draw": false,
		"pause_votes": {},
		"ready_ids": [],
		"spectator_ids": []
	}

# Placeholder implementations for game endpoints
func _join_game_session(session_id: String, user_data: Dictionary, data: Dictionary) -> Dictionary:
	return {"status_code": 200, "headers": {"Content-Type": "application/json"}, "body": JSON.stringify({"message": "Joined game session"})}

func _start_game_session(session_id: String, user_data: Dictionary) -> Dictionary:
	return {"status_code": 200, "headers": {"Content-Type": "application/json"}, "body": JSON.stringify({"message": "Game session started"})}

func _toggle_player_ready(session_id: String, user_data: Dictionary) -> Dictionary:
	return {"status_code": 200, "headers": {"Content-Type": "application/json"}, "body": JSON.stringify({"message": "Player ready status toggled"})}

func _get_game_session(session_id: String, user_data: Dictionary) -> Dictionary:
	return {"status_code": 200, "headers": {"Content-Type": "application/json"}, "body": JSON.stringify({"message": "Game session state"})}

func _submit_move(session_id: String, user_data: Dictionary, data: Dictionary) -> Dictionary:
	return {"status_code": 200, "headers": {"Content-Type": "application/json"}, "body": JSON.stringify({"message": "Move submitted"})}

func _forfeit_game(session_id: String, user_data: Dictionary) -> Dictionary:
	return {"status_code": 200, "headers": {"Content-Type": "application/json"}, "body": JSON.stringify({"message": "Game forfeited"})}

func _get_cards() -> Dictionary:
	# Get all cards from the game logic engine
	var cards = game_logic.get_all_cards()
	
	# Sort cards by strength (strongest to weakest) like the original Go backend
	cards.sort_custom(func(a, b): return a.strength > b.strength)
	
	return {
		"status_code": 200,
		"headers": {"Content-Type": "application/json"},
		"body": JSON.stringify(cards)
	}

func _get_card(card_id: String) -> Dictionary:
	var card = game_logic.get_card_data(card_id)
	
	if card.is_empty():
		return {
			"status_code": 404,
			"headers": {"Content-Type": "application/json"},
			"body": JSON.stringify({"error": "Card not found"})
		}
	
	return {
		"status_code": 200,
		"headers": {"Content-Type": "application/json"},
		"body": JSON.stringify(card)
	}

func _resolve_challenge(challenge_id: String, user_data: Dictionary, data: Dictionary) -> Dictionary:
	return {"status_code": 200, "headers": {"Content-Type": "application/json"}, "body": JSON.stringify({"message": "Challenge resolved"})}

func _cleanup_expired_sessions():
	var current_time = Time.get_unix_time_from_system()
	var timeout_minutes = ConfigManager.get_value("game", "session_timeout_minutes", 30)
	var timeout_seconds = timeout_minutes * 60
	
	for session_id in game_sessions:
		var session = game_sessions[session_id]
		var created_time = Time.get_unix_time_from_datetime_string(session.created_at)
		if current_time - created_time > timeout_seconds:
			game_sessions.erase(session_id)
			print("Cleaned up expired session: " + session_id)

func _on_database_connected():
	print("Database connection established")

func _on_database_failed(error: String):
	print("Database connection failed: " + error) 
