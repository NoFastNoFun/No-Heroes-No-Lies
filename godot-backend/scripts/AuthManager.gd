extends Node

# Authentication manager for JWT token handling
# Replaces the Go JWT authentication system

signal user_authenticated(user_data: Dictionary)
signal authentication_failed(error: String)
signal token_refreshed(new_token: String)

var jwt_secret: String
var token_expiry_hours: int
var refresh_token_expiry_days: int

# In-memory token storage (in production, consider using a more secure storage)
var active_tokens: Dictionary = {}

func _ready():
	# Wait for config to load
	await ConfigManager.config_loaded
	
	jwt_secret = ConfigManager.get_jwt_secret()
	token_expiry_hours = ConfigManager.get_value("auth", "token_expiry_hours", 24)
	refresh_token_expiry_days = ConfigManager.get_value("auth", "refresh_token_expiry_days", 7)

# Generate JWT token
func generate_token(user_data: Dictionary) -> String:
	var header = {
		"alg": "HS256",
		"typ": "JWT"
	}
	
	var payload = {
		"user_id": user_data.get("id", ""),
		"email": user_data.get("email", ""),
		"exp": Time.get_unix_time_from_system() + (token_expiry_hours * 3600),
		"iat": Time.get_unix_time_from_system()
	}
	
	return _create_jwt(header, payload)

# Generate refresh token
func generate_refresh_token(user_id: String) -> String:
	var header = {
		"alg": "HS256",
		"typ": "JWT"
	}
	
	var payload = {
		"user_id": user_id,
		"type": "refresh",
		"exp": Time.get_unix_time_from_system() + (refresh_token_expiry_days * 86400),
		"iat": Time.get_unix_time_from_system()
	}
	
	return _create_jwt(header, payload)

# Validate JWT token
func validate_token(token: String) -> Dictionary:
	if token == "":
		return {"valid": false, "error": "Empty token"}
	
	var parts = token.split(".")
	if parts.size() != 3:
		return {"valid": false, "error": "Invalid token format"}
	
	var header_json = Marshalls.base64_to_utf8(parts[0])
	var payload_json = Marshalls.base64_to_utf8(parts[1])
	var signature = parts[2]
	
	var header = JSON.parse_string(header_json)
	var payload = JSON.parse_string(payload_json)
	
	if not header or not payload:
		return {"valid": false, "error": "Invalid token structure"}
	
	# Check if token is expired
	var current_time = Time.get_unix_time_from_system()
	if payload.has("exp") and payload.exp < current_time:
		return {"valid": false, "error": "Token expired"}
	
	# Verify signature (simplified - in production use proper crypto)
	var expected_signature = _generate_signature(parts[0] + "." + parts[1])
	if signature != expected_signature:
		return {"valid": false, "error": "Invalid signature"}
	
	return {"valid": true, "user_data": payload}

# Extract user data from token
func get_user_from_token(token: String) -> Dictionary:
	var validation = validate_token(token)
	if validation.valid:
		return validation.user_data
	return {}

# Check if user is authenticated
func is_authenticated(token: String) -> bool:
	var validation = validate_token(token)
	return validation.valid

# Refresh token
func refresh_token(refresh_token: String) -> Dictionary:
	var validation = validate_token(refresh_token)
	if not validation.valid:
		return {"success": false, "error": validation.error}
	
	if validation.user_data.get("type") != "refresh":
		return {"success": false, "error": "Not a refresh token"}
	
	# Generate new access token
	var user_data = {
		"id": validation.user_data.get("user_id", ""),
		"email": validation.user_data.get("email", "")
	}
	
	var new_token = generate_token(user_data)
	var new_refresh_token = generate_refresh_token(validation.user_data.get("user_id", ""))
	
	token_refreshed.emit(new_token)
	
	return {
		"success": true,
		"access_token": new_token,
		"refresh_token": new_refresh_token
	}

# Login user
func login_user(email: String, password: String) -> void:
	var auth_data = {
		"identity": email,
		"password": password
	}
	
	# This would typically make an async call to the database
	# For now, we'll simulate the authentication
	Database.authenticate_user(email, password)
	
	# In a real implementation, you'd wait for the database response
	# and then generate tokens based on the user data

# Register user
func register_user(email: String, password: String, password_confirm: String) -> void:
	if password != password_confirm:
		authentication_failed.emit("Passwords do not match")
		return
	
	var user_data = {
		"email": email,
		"password": password,
		"passwordConfirm": password_confirm
	}
	
	Database.create_user(email, password, password_confirm)

# Logout user
func logout_user(token: String) -> void:
	# In a real implementation, you might want to blacklist the token
	# For now, we'll just remove it from active tokens
	if active_tokens.has(token):
		active_tokens.erase(token)

# Helper function to create JWT
func _create_jwt(header: Dictionary, payload: Dictionary) -> String:
	var header_json = JSON.stringify(header)
	var payload_json = JSON.stringify(payload)
	
	var header_b64 = Marshalls.utf8_to_base64(header_json)
	var payload_b64 = Marshalls.utf8_to_base64(payload_json)
	
	var data = header_b64 + "." + payload_b64
	var signature = _generate_signature(data)
	
	return data + "." + signature

# Helper function to generate signature
func _generate_signature(data: String) -> String:
	# This is a simplified signature generation
	# In production, you should use proper HMAC-SHA256
	var hmac = HashingContext.new()
	hmac.start(HashingContext.HASH_SHA256)
	hmac.update(data.to_utf8_buffer())
	hmac.update(jwt_secret.to_utf8_buffer())
	var hash = hmac.finish()
	return Marshalls.raw_to_base64(hash).replace("+", "-").replace("/", "_").replace("=", "")

# Extract token from Authorization header
func extract_token_from_header(auth_header: String) -> String:
	if auth_header.begins_with("Bearer "):
		return auth_header.substr(7)
	return ""

# Middleware function for protecting routes
func require_auth(token: String) -> Dictionary:
	if token == "":
		return {"authorized": false, "error": "No token provided"}
	
	var validation = validate_token(token)
	if not validation.valid:
		return {"authorized": false, "error": validation.error}
	
	return {"authorized": true, "user_data": validation.user_data} 
