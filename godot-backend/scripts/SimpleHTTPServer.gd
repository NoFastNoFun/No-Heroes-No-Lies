extends Node

# Simple HTTP Server implementation using Godot's built-in functionality
# This provides basic HTTP server capabilities for the game server

signal request_received(request: HTTPRequest)

var server: TCPServer
var is_listening: bool = false
var port: int = 8080
var host: String = "localhost"
var active_connections: Array[StreamPeerTCP] = []

func listen(server_port: int, server_host: String) -> int:
	port = server_port
	host = server_host
	
	server = TCPServer.new()
	var result = server.listen(port, host)
	
	if result == OK:
		is_listening = true
		print("HTTP Server listening on " + host + ":" + str(port))
		# Start processing connections
		_process_connections()
	else:
		print("Failed to start HTTP server: " + str(result))
	
	return result

func stop():
	if server:
		server.stop()
		is_listening = false
		print("HTTP Server stopped")

func _process_connections():
	while is_listening:
		# Check for new connections
		if server.is_connection_available():
			var connection = server.take_connection()
			active_connections.append(connection)
			_handle_connection(connection)
		
		# Process existing connections
		for i in range(active_connections.size() - 1, -1, -1):
			var connection = active_connections[i]
			if connection.get_status() != StreamPeerTCP.STATUS_CONNECTED:
				active_connections.remove_at(i)
				continue
			
			# Check if there's data to read
			if connection.get_available_bytes() > 0:
				_process_connection_data(connection)
		
		# Small delay to prevent busy waiting
		await get_tree().create_timer(0.01).timeout

func _handle_connection(connection: StreamPeerTCP):
	# Connection is already added to active_connections
	# Data will be processed in _process_connections
	pass

func _process_connection_data(connection: StreamPeerTCP):
	# Read the HTTP request
	var request_data = await _read_http_request(connection)
	if request_data.is_empty():
		connection.disconnect_from_host()
		return
	
	# Create a proper HTTPRequest object
	var http_request = _create_http_request(request_data, connection)
	
	# Emit the signal
	request_received.emit(http_request)

func _read_http_request(connection: StreamPeerTCP) -> Dictionary:
	var request_text = ""
	var buffer = PackedByteArray()
	
	# Read the request with timeout
	var timeout = 0
	var max_timeout = 1000  # 1 second timeout
	
	while connection.get_available_bytes() > 0 and timeout < max_timeout:
		var chunk = connection.get_data(1024)
		if chunk[0] == OK:
			buffer.append_array(chunk[1])
			timeout = 0  # Reset timeout on successful read
		else:
			timeout += 1
			await get_tree().create_timer(0.001).timeout
	
	request_text = buffer.get_string_from_utf8()
	
	# Parse the HTTP request
	return _parse_http_request(request_text)

func _parse_http_request(request_text: String) -> Dictionary:
	var lines = request_text.split("\n")
	if lines.size() == 0:
		return {}
	
	# Parse the first line (method, path, version)
	var first_line = lines[0].strip_edges()
	var parts = first_line.split(" ")
	if parts.size() < 2:
		return {}
	
	var method = parts[0]
	var path = parts[1]
	
	# Parse headers
	var headers = {}
	var body = ""
	var in_body = false
	
	for i in range(1, lines.size()):
		var line = lines[i].strip_edges()
		if line == "":
			in_body = true
			continue
		
		if in_body:
			body += line + "\n"
		else:
			var colon_pos = line.find(":")
			if colon_pos > 0:
				var key = line.substr(0, colon_pos).strip_edges()
				var value = line.substr(colon_pos + 1).strip_edges()
				headers[key] = value
	
	return {
		"method": method,
		"path": path,
		"headers": headers,
		"body": body.strip_edges()
	}

func _create_http_request(request_data: Dictionary, connection: StreamPeerTCP) -> HTTPRequest:
	# Create a custom HTTPRequest-like object
	var http_request = HTTPRequest.new()
	
	# Store the connection for sending responses
	http_request.set_meta("connection", connection)
	http_request.set_meta("method", request_data.get("method", "GET"))
	http_request.set_meta("path", request_data.get("path", "/"))
	http_request.set_meta("headers", request_data.get("headers", {}))
	http_request.set_meta("body", request_data.get("body", ""))
	
	# Add methods to the request object
	http_request.get_method = func(): return http_request.get_meta("method")
	http_request.get_path = func(): return http_request.get_meta("path")
	http_request.get_headers = func(): return http_request.get_meta("headers")
	http_request.get_body = func(): return http_request.get_meta("body").to_utf8_buffer()
	
	# Add response methods
	http_request.send_response = func(status_code: int, headers: Dictionary):
		_send_http_response(http_request, status_code, headers)
	
	http_request.send_body = func(body: PackedByteArray):
		_send_http_body(http_request, body)
	
	return http_request

func _send_http_response(request: HTTPRequest, status_code: int, headers: Dictionary):
	var connection = request.get_meta("connection")
	if not connection or connection.get_status() != StreamPeerTCP.STATUS_CONNECTED:
		return
	
	var status_text = "OK"
	match status_code:
		200: status_text = "OK"
		201: status_text = "Created"
		400: status_text = "Bad Request"
		401: status_text = "Unauthorized"
		404: status_text = "Not Found"
		500: status_text = "Internal Server Error"
		_: status_text = "Unknown"
	
	var response = "HTTP/1.1 " + str(status_code) + " " + status_text + "\r\n"
	
	# Add headers
	for key in headers:
		response += key + ": " + str(headers[key]) + "\r\n"
	
	response += "\r\n"
	
	# Send the response
	var response_bytes = response.to_utf8_buffer()
	connection.put_data(response_bytes)

func _send_http_body(request: HTTPRequest, body: PackedByteArray):
	var connection = request.get_meta("connection")
	if not connection or connection.get_status() != StreamPeerTCP.STATUS_CONNECTED:
		return
	
	# Send the body
	connection.put_data(body)
	
	# Close the connection after sending
	connection.disconnect_from_host() 
