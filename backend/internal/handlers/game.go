package handlers

import (
	"encoding/json"
	"net/http"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/models"
	"no-heroes-no-lies/internal/services"
	"no-heroes-no-lies/internal/view"

	"github.com/go-chi/chi/v5"
)

// RegisterGameRoutes registers HTTP endpoints under /game.
func RegisterGameRoutes(r chi.Router, svc *services.GameService) {
	r.Route("/game", func(r chi.Router) {
		r.Get("/{id}", getSession(svc))
		r.Post("/{id}/move", postMove(svc))
		r.Post("/{id}/forfeit", postForfeit(svc))
	})
}

// @Summary Get game session
// @Description Get current game session state for the authenticated player
// @Tags game
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Security BearerAuth
// @Success 200 {object} view.GameView
// @Failure 401 {string} string "Unauthorized"
// @Failure 502 {string} string "Bad Gateway"
// @Router /game/{id} [get]
func getSession(svc *services.GameService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID, err := auth.PlayerIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		id := chi.URLParam(r, "id")
		session, err := svc.FetchSession(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		resp := view.Build(session, playerID)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// @Summary Make a move
// @Description Submit a move in the game (demask, fight, or power)
// @Tags game
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Param move body models.MovePayload true "Move details"
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Router /game/{id}/move [post]
func postMove(svc *services.GameService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")
		playerID, err := auth.PlayerIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var payload models.MovePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if err := svc.ApplyMove(sessionID, playerID, payload); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary Forfeit game
// @Description Forfeit the current game session
// @Tags game
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Router /game/{id}/forfeit [post]
func postForfeit(svc *services.GameService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		playerID, err := auth.PlayerIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		sessionID := chi.URLParam(r, "id")
		if err := svc.Forfeit(sessionID, playerID); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
