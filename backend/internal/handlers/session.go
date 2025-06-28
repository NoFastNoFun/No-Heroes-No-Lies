package handlers

import (
	"encoding/json"
	"net/http"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/services"

	"github.com/go-chi/chi/v5"
)

// RegisterSessionRoutes mounts session creation / join / start.
func RegisterSessionRoutes(r chi.Router, sess *services.SessionService) {
	r.Post("/game", createSession(sess))
	r.Post("/game/{id}/join", joinSession(sess))
	r.Post("/game/{id}/start", startSession(sess))
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
		playerID, err := auth.PlayerIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		sessionID := chi.URLParam(r, "id")

		session, err := sess.JoinSession(sessionID, playerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(session)
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
