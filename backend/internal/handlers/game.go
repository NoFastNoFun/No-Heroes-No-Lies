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
	})
}

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

func postMove(svc *services.GameService) http.HandlerFunc {
	type req struct {
		PowerID string `json:"power_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")

		playerID, err := auth.PlayerIDFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var body req
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		power := models.Power{ID: body.PowerID}
		if err := svc.ApplyMove(sessionID, power, playerID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
