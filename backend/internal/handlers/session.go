package handlers

import (
	"encoding/json"
	"net/http"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/services"

	"github.com/go-chi/chi/v5"
)

// RegisterSessionRoutes mounts session creation / join / start.
func RegisterSessionRoutes(r chi.Router, svc *services.SessionService) {
	r.Post("/game", createSession(svc))
	r.Post("/game/{id}/join", joinSession(svc))
	r.Post("/game/{id}/start", startSession(svc))
	r.Post("/game/{id}/ready", readyToggle(svc))
}

func readyToggle(sess *services.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID, _ := auth.PlayerIDFromContext(r.Context())
		sessionID := chi.URLParam(r, "id")
		sn, err := sess.ToggleReady(sessionID, playerID)
		if err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		_ = json.NewEncoder(w).Encode(sn)
	}
}

func createSession(sess *services.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID, err := auth.PlayerIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		session, err := sess.CreateSession(playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		_ = json.NewEncoder(w).Encode(session)
	}
}

func joinSession(sess *services.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID, _ := auth.PlayerIDFromContext(r.Context())
		sessionID := chi.URLParam(r, "id")
		spec := r.URL.Query().Get("spectator") == "1"

		sn, err := sess.JoinSession(sessionID, playerID, spec)
		if err != nil {
			http.Error(w, err.Error(), 403)
			return
		}
		_ = json.NewEncoder(w).Encode(sn)
	}
}

func startSession(sess *services.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")

		session, err := sess.StartSession(sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(session)
	}
}
