package games

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/service"
	"github.com/no-heroes-no-lies/backend/internal/transport/http/middleware"
	"github.com/no-heroes-no-lies/backend/pkg/response"
)

func State(stateService *service.GameStateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.JSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		playerID, ok := middleware.GetPlayerID(r.Context())
		if !ok {
			response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		idStr := r.PathValue("id")
		if idStr == "" {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "missing game id"})
			return
		}
		gameID, err := uuid.Parse(idStr)
		if err != nil {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid game id"})
			return
		}
		view, err := stateService.GetState(r.Context(), gameID, playerID)
		if err != nil {
			if err == service.ErrNotFound {
				response.JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
				return
			}
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		response.JSON(w, http.StatusOK, view)
	}
}
