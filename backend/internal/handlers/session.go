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

// @Summary Toggle player ready status
// @Description Toggle the ready status of a player in a session
// @Tags session
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Security BearerAuth
// @Success 200 {object} models.GameSession
// @Failure 403 {string} string "Forbidden"
// @Router /game/{id}/ready [post]
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

// @Summary Create new game session
// @Description Create a new game session for the authenticated player
// @Tags session
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.GameSession
// @Failure 401 {string} string "Unauthorized"
// @Failure 502 {string} string "Bad Gateway"
// @Router /game [post]
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

// @Summary Join game session
// @Description Join an existing game session as player or spectator
// @Tags session
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param spectator query string false "Set to '1' to join as spectator"
// @Security BearerAuth
// @Success 200 {object} models.GameSession
// @Failure 403 {string} string "Forbidden"
// @Router /game/{id}/join [post]
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

// @Summary Start game session
// @Description Start a game session when all players are ready
// @Tags session
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} models.GameSession
// @Failure 400 {string} string "Bad Request"
// @Router /game/{id}/start [post]
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
