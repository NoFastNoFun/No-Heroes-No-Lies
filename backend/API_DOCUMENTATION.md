# API Documentation

This backend provides comprehensive OpenAPI documentation for all routes and implements graceful shutdown for production deployments.

## Graceful Shutdown

The server implements graceful shutdown that:

- Traps SIGINT and SIGTERM signals
- Waits up to 10 seconds for in-flight requests to complete
- Shuts down cleanly - perfect for Coolify rolling deploys
- Logs shutdown progress for monitoring

### Testing Graceful Shutdown

**Linux/macOS:**

```bash
# Build and test
go build ./cmd/server
./test_graceful_shutdown.sh
```

**Windows:**

```cmd
# Build and test
go build ./cmd/server
test_graceful_shutdown.bat
```

## Accessing the Documentation

Once the server is running, you can access the OpenAPI documentation at:

- **Swagger UI**: `http://localhost:8080/swagger/`
- **OpenAPI JSON**: `http://localhost:8080/docs/swagger.json`

## API Endpoints

### Health Check

- `GET /api/health` - Check if the service is running

### Session Management

- `POST /api/game` - Create new game session
- `POST /api/game/{id}/join` - Join game session (with optional spectator parameter)
- `POST /api/game/{id}/start` - Start game session
- `POST /api/game/{id}/ready` - Toggle player ready status

### Game Actions

- `GET /api/game/{id}` - Get current game session state
- `POST /api/game/{id}/move` - Submit a move (demask, fight, or power)
- `POST /api/game/{id}/forfeit` - Forfeit the current game
- `POST /api/game/{id}/challenge` - Resolve a challenge

## Authentication

All endpoints (except health check) require Bearer token authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-token>
```

## Data Models

The API uses several key data models:

- **MovePayload**: Contains move details for game actions
- **GameSession**: Complete game session information
- **SessionView**: Player-specific view of the game state
- **Power**: Game power/ability definitions
- **Loot**: Monster reward information

## Examples

### Creating a Session

```bash
curl -X POST http://localhost:8080/api/game \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

### Making a Move

```bash
curl -X POST http://localhost:8080/api/game/session123/move \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "demask",
    "target_player": "player456",
    "guess": "hero_name"
  }'
```

## Error Responses

The API returns appropriate HTTP status codes:

- `200` - Success
- `204` - Success (no content)
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `502` - Bad Gateway

For detailed error information, refer to the OpenAPI specification.
