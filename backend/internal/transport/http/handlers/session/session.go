package session

import (
	"encoding/json"
	"net/http"

	"github.com/no-heroes-no-lies/backend/internal/service"
	"github.com/no-heroes-no-lies/backend/pkg/response"
)

type CreateRequest struct {
	Username string `json:"username"`
}

func Create(sessionService *service.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.JSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
			return
		}
		_, out, err := sessionService.CreateSession(r.Context(), req.Username)
		if err != nil {
			if err == service.ErrInvalidInput {
				response.JSON(w, http.StatusBadRequest, map[string]string{"error": "username required"})
				return
			}
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		response.JSON(w, http.StatusCreated, out)
	}
}
