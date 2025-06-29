extends Node

# Configuration manager for the No Heroes No Lies server
# Replaces the Go config system

signal config_loaded

var config_data: Dictionary = {}

# Default configuration values
var defaults: Dictionary = {
	"server": {
		"port": 8080,
		"host": "0.0.0.0",
		"max_players": 8,
		"graceful_shutdown_timeout": 10.0
	},
	"database": {
		"pocketbase_url": "http://localhost:8090",
		"pocketbase_admin_key": "",
		"connection_timeout": 30.0
	},
	"auth": {
		"jwt_secret": "your-secret-key-change-in-production",
		"token_expiry_hours": 24,
		"refresh_token_expiry_days": 7
	},
	"cors": {
		"allowed_origins": ["http://localhost:3000", "http://localhost:8080"],
		"allowed_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
		"allowed_headers": ["Accept", "Authorization", "Content-Type", "X-CSRF-Token"]
	},
	"game": {
		"session_timeout_minutes": 30,
		"move_challenge_window_seconds": 5,
		"max_spectators": 4
	}
}

func _ready():
	load_config()

func load_config():
	# Try to load from config file first
	var config_file = FileAccess.open("res://config.json", FileAccess.READ)
	if config_file:
		var json_string = config_file.get_as_text()
		config_file.close()
		
		var json = JSON.new()
		var parse_result = json.parse(json_string)
		
		if parse_result == OK:
			config_data = json.data
			print("Configuration loaded from config.json")
		else:
			print("Failed to parse config.json, using defaults")
			config_data = defaults.duplicate(true)
	else:
		print("No config.json found, using defaults")
		config_data = defaults.duplicate(true)
	
	# Merge with defaults to ensure all required keys exist
	merge_with_defaults()
	
	config_loaded.emit()

func merge_with_defaults():
	for section in defaults:
		if not config_data.has(section):
			config_data[section] = defaults[section].duplicate()
		else:
			for key in defaults[section]:
				if not config_data[section].has(key):
					config_data[section][key] = defaults[section][key]

func get_value(section: String, key: String, default_value = null):
	if config_data.has(section) and config_data[section].has(key):
		return config_data[section][key]
	return default_value

func set_value(section: String, key: String, value):
	if not config_data.has(section):
		config_data[section] = {}
	config_data[section][key] = value

func save_config():
	var config_file = FileAccess.open("res://config.json", FileAccess.WRITE)
	if config_file:
		var json_string = JSON.stringify(config_data, "\t")
		config_file.store_string(json_string)
		config_file.close()
		print("Configuration saved to config.json")
	else:
		print("Failed to save configuration")

# Convenience getters for common config values
func get_server_port() -> int:
	return get_value("server", "port", 8080)

func get_server_host() -> String:
	return get_value("server", "host", "0.0.0.0")

func get_pocketbase_url() -> String:
	return get_value("database", "pocketbase_url", "http://localhost:8090")

func get_pocketbase_admin_key() -> String:
	return get_value("database", "pocketbase_admin_key", "")

func get_jwt_secret() -> String:
	return get_value("auth", "jwt_secret", "your-secret-key-change-in-production")

func get_allowed_origins() -> Array:
	return get_value("cors", "allowed_origins", ["http://localhost:3000"])

func get_graceful_shutdown_timeout() -> float:
	return get_value("server", "graceful_shutdown_timeout", 10.0) 