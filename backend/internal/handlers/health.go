package handlers

import (
	"net/http"
)

// HealthHandler responds with 200 OK for liveness checks.
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
