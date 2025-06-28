package handlers

import (
	"net/http"
	"time"
)

// HealthHandler responds with 200 OK for liveness checks.
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Accept json
// @Produce plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request context has been cancelled (shutdown in progress)
	select {
	case <-r.Context().Done():
		// Context cancelled, server is shutting down
		http.Error(w, "Server shutting down", http.StatusServiceUnavailable)
		return
	default:
		// Normal health check
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}

// SlowHealthHandler simulates a long-running request to test graceful shutdown.
// This endpoint takes 5 seconds to respond, useful for testing shutdown behavior.
// @Summary Slow health check
// @Description Simulates a long-running request (5 seconds) to test graceful shutdown behavior
// @Tags health
// @Accept json
// @Produce plain
// @Success 200 {string} string "Slow health check completed"
// @Failure 503 {string} string "Server shutting down"
// @Router /health/slow [get]
func SlowHealthHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate a long-running operation
	select {
	case <-time.After(5 * time.Second):
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Slow health check completed"))
	case <-r.Context().Done():
		// Request cancelled due to server shutdown
		http.Error(w, "Request cancelled - server shutting down", http.StatusServiceUnavailable)
	}
}
