# Graceful Shutdown Implementation

This server implements production-ready graceful shutdown that's perfect for containerized deployments and rolling updates.

## Features

### ✅ Signal Handling

- **SIGINT** (Ctrl+C) - Interrupt signal
- **SIGTERM** - Termination signal (default for `kill` command)
- Ignores SIGKILL (can't be caught anyway)

### ✅ Request Completion

- Waits up to **10 seconds** for in-flight requests to complete
- Uses Go's `http.Server.Shutdown()` method
- Context-based timeout ensures clean shutdown

### ✅ Production Ready

- Perfect for **Coolify rolling deploys**
- Works with **Docker containers**
- Compatible with **Kubernetes** pod termination
- Logs shutdown progress for monitoring

## How It Works

```go
// 1. Start server in goroutine
go func() {
    srv.ListenAndServe()
}()

// 2. Wait for shutdown signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// 3. Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

## Testing

### Manual Testing

**Start the server:**

```bash
go run cmd/server/main.go
```

**In another terminal, test graceful shutdown:**

```bash
# Test normal health check
curl http://localhost:8080/api/health

# Test slow endpoint (5 seconds)
curl http://localhost:8080/api/health/slow

# Send SIGTERM to trigger shutdown
kill -TERM <PID>
```

### Automated Testing

**Linux/macOS:**

```bash
chmod +x test_graceful_shutdown.sh
./test_graceful_shutdown.sh
```

**Windows:**

```cmd
test_graceful_shutdown.bat
```

## Expected Behavior

### Normal Shutdown

```
2024/01/01 12:00:00 server listening on :8080
2024/01/01 12:00:05 shutting down server...
2024/01/01 12:00:05 server exited
```

### With In-Flight Requests

```
2024/01/01 12:00:00 server listening on :8080
2024/01/01 12:00:05 shutting down server...
# Server waits for requests to complete (up to 10s)
2024/01/01 12:00:10 server exited
```

### Forced Shutdown

```
2024/01/01 12:00:00 server listening on :8080
2024/01/01 12:00:05 shutting down server...
2024/01/01 12:00:15 server forced to shutdown: context deadline exceeded
```

## Deployment Considerations

### Docker

```dockerfile
# Use tini for proper signal handling
RUN apk add --no-cache tini
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./server"]
```

### Kubernetes

```yaml
spec:
  containers:
  - name: server
    image: your-app:latest
    lifecycle:
      preStop:
        exec:
          command: ["/bin/sh", "-c", "sleep 5"]
    terminationGracePeriodSeconds: 15
```

### Coolify

- Set **grace period** to 15 seconds
- Enable **rolling updates**
- Server will handle SIGTERM gracefully

## Monitoring

The server logs key events:

- `server listening on :8080` - Server started
- `shutting down server...` - Shutdown initiated
- `server exited` - Clean shutdown
- `server forced to shutdown: ...` - Timeout exceeded

## Troubleshooting

### Server Won't Shutdown

- Check if requests are stuck in middleware
- Verify no blocking operations in handlers
- Ensure database connections are properly closed

### Shutdown Too Slow

- Reduce timeout from 10 seconds
- Optimize slow database queries
- Review long-running operations

### Shutdown Too Fast

- Increase timeout beyond 10 seconds
- Check for proper signal handling
- Verify container runtime settings
