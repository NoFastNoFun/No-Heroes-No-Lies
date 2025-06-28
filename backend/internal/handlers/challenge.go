package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/services"

	"github.com/go-chi/chi/v5"
)

// RegisterChallengeRoute mounts POST /game/{id}/challenge.
func RegisterChallengeRoute(r chi.Router, svc *services.GameService) {
	r.Post("/game/{id}/challenge", challengeHandler(svc))
}

// @Summary Resolve challenge
// @Description Resolve a challenge in the game
// @Tags game
// @Accept json
// @Produce json
// @Param id path string true "Session ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden"
// @Router /game/{id}/challenge [post]
func challengeHandler(svc *services.GameService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")
		challengerID, err := auth.PlayerIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err := svc.ResolveChallenge(sessionID, challengerID, time.Now().UnixMilli()); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(struct {
			Result string `json:"result"`
		}{Result: "resolved"})
	}
}
