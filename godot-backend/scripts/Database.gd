extends Node

# Database manager for PocketBase integration
# Replaces the Go pb client functionality

signal connection_established
signal connection_failed(error: String)
signal cards_loaded(cards: Array)
signal powers_loaded(powers: Array)
signal session_loaded(session: Dictionary)
signal session_created(session: Dictionary)
signal session_updated(session_id: String)
signal user_authenticated(user_data: Dictionary)
signal user_registered(user_data: Dictionary)

var http_request: HTTPRequest
var base_url: String
var admin_key: String
var is_connected: bool = false

# Request tracking for async operations
var pending_requests: Dictionary = {}

func _ready():
	http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(_on_request_completed)
	
	# Wait for config to load
	await ConfigManager.config_loaded
	
	base_url = ConfigManager.get_pocketbase_url()
	admin_key = ConfigManager.get_pocketbase_admin_key()
	
	# Test connection
	test_connection()

func test_connection():
	var headers = ["Content-Type: application/json"]
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var url = base_url + "/api/health"
	var error = http_request.request(url, headers, HTTPClient.METHOD_GET)
	
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func _on_request_completed(result: int, response_code: int, headers: PackedStringArray, body: PackedByteArray):
	if result != HTTPRequest.RESULT_SUCCESS:
		is_connected = false
		connection_failed.emit("Request failed: " + str(result))
		return
	
	var response_body = body.get_string_from_utf8()
	var json = JSON.new()
	var parse_result = json.parse(response_body)
	
	if parse_result != OK:
		is_connected = false
		connection_failed.emit("Failed to parse response")
		return
	
	var data = json.data
	
	# Handle different response types based on the request
	# This is a simplified version - in a real implementation you'd track the specific request
	if response_code == 200 or response_code == 201:
		# Success - emit appropriate signal based on the data structure
		if data.has("items"):
			# List response (cards, powers, etc.)
			if data.items.size() > 0 and data.items[0].has("type"):
				if data.items[0].type == "hero" or data.items[0].type == "monster":
					cards_loaded.emit(data.items)
				else:
					powers_loaded.emit(data.items)
		elif data.has("id") and data.has("player_ids"):
			# Game session response
			if data.has("created"):
				session_created.emit(data)
			else:
				session_loaded.emit(data)
		elif data.has("token"):
			# Authentication response
			user_authenticated.emit(data)
		elif data.has("id") and data.has("email"):
			# User creation response
			user_registered.emit(data)
		elif data.has("code") == false:  # Health check response
			is_connected = true
			connection_established.emit()
	else:
		is_connected = false
		connection_failed.emit("HTTP " + str(response_code) + ": " + response_body)

# Generic database operations with proper async handling
func get_records(collection: String, filter: String = "", expand: String = "") -> void:
	var url = base_url + "/api/collections/" + collection + "/records"
	
	if filter != "":
		url += "?filter=" + filter.uri_encode()
		if expand != "":
			url += "&expand=" + expand.uri_encode()
	elif expand != "":
		url += "?expand=" + expand.uri_encode()
	
	var headers = ["Content-Type: application/json"]
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "get_records",
		"collection": collection,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_GET)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func create_record(collection: String, data: Dictionary) -> void:
	var url = base_url + "/api/collections/" + collection + "/records"
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var json_string = JSON.stringify(data)
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "create_record",
		"collection": collection,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_POST, json_string)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func update_record(collection: String, record_id: String, data: Dictionary) -> void:
	var url = base_url + "/api/collections/" + collection + "/records/" + record_id
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var json_string = JSON.stringify(data)
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "update_record",
		"collection": collection,
		"record_id": record_id,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_PATCH, json_string)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func delete_record(collection: String, record_id: String) -> void:
	var url = base_url + "/api/collections/" + collection + "/records/" + record_id
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "delete_record",
		"collection": collection,
		"record_id": record_id,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_DELETE)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

# Game-specific database operations
func get_game_session(session_id: String) -> void:
	var url = base_url + "/api/collections/game_sessions/records/" + session_id
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "get_game_session",
		"session_id": session_id,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_GET)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func create_game_session(player_ids: Array) -> void:
	var data = {
		"player_ids": player_ids,
		"state": {},
		"is_active": true
	}
	
	var url = base_url + "/api/collections/game_sessions/records"
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var json_string = JSON.stringify(data)
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "create_game_session",
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_POST, json_string)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func update_game_session(session_id: String, state: Dictionary) -> void:
	var data = {
		"state": state,
		"updated_at": Time.get_datetime_string_from_system()
	}
	
	var url = base_url + "/api/collections/game_sessions/records/" + session_id
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var json_string = JSON.stringify(data)
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "update_game_session",
		"session_id": session_id,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_PATCH, json_string)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func update_session_players(session_id: String, player_ids: Array) -> void:
	var data = {
		"player_ids": player_ids
	}
	
	var url = base_url + "/api/collections/game_sessions/records/" + session_id
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var json_string = JSON.stringify(data)
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "update_session_players",
		"session_id": session_id,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_PATCH, json_string)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

# Cards and Powers operations
func get_cards(card_type: String = "") -> void:
	var filter = ""
	if card_type != "":
		filter = "(type='" + card_type + "')"
	
	var url = base_url + "/api/collections/cards/records?perPage=200"
	if filter != "":
		url += "&filter=" + filter.uri_encode()
	
	var headers = ["Content-Type: application/json"]
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "get_cards",
		"card_type": card_type,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_GET)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func get_card(card_id: String) -> void:
	var url = base_url + "/api/collections/cards/records/" + card_id
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "get_card",
		"card_id": card_id,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_GET)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func get_powers() -> void:
	var url = base_url + "/api/collections/powers/records?perPage=200"
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "get_powers",
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_GET)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func get_power(power_id: String) -> void:
	var url = base_url + "/api/collections/powers/records/" + power_id
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "get_power",
		"power_id": power_id,
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_GET)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

# Authentication operations
func authenticate_user(email: String, password: String) -> void:
	var url = base_url + "/api/collections/games_accounts/auth-with-password"
	var headers = ["Content-Type: application/json"]
	
	var data = {
		"identity": email,
		"password": password
	}
	
	var json_string = JSON.stringify(data)
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "authenticate_user",
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_POST, json_string)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func create_user(email: String, password: String, password_confirm: String, username: String = "", display_name: String = "") -> void:
	var url = base_url + "/api/collections/games_accounts/records"
	var headers = ["Content-Type: application/json"]
	
	var data = {
		"email": email,
		"password": password,
		"passwordConfirm": password_confirm
	}
	
	if username != "":
		data["username"] = username
	if display_name != "":
		data["display_name"] = display_name
	
	var json_string = JSON.stringify(data)
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "create_user",
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_POST, json_string)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

func verify_user_token(token: String) -> void:
	var url = base_url + "/api/collections/games_accounts/auth-verify"
	var headers = ["Content-Type: application/json"]
	
	if admin_key != "":
		headers.append("AO_KEY: " + admin_key)
	headers.append("Authorization: Bearer " + token)
	
	var request_id = str(randi())
	pending_requests[request_id] = {
		"type": "verify_user_token",
		"url": url
	}
	
	var error = http_request.request(url, headers, HTTPClient.METHOD_POST)
	if error != OK:
		connection_failed.emit("Failed to initiate request")

# File operations
func get_file_url(collection_id: String, record_id: String, filename: String) -> String:
	return base_url + "/api/files/" + collection_id + "/" + record_id + "/" + filename 