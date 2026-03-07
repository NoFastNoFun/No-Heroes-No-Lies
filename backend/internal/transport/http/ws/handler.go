package ws

import (
	"net/http"

	"github.com/google/uuid"
)

func Handler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "missing game id", http.StatusBadRequest)
			return
		}
		gameID, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "invalid game id", http.StatusBadRequest)
			return
		}
		hub.HandleWS(w, r, gameID)
	}
}
