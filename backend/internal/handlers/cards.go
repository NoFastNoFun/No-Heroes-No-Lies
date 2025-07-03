package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"no-heroes-no-lies/internal/db"

	"github.com/go-chi/chi/v5"
)

// CardsHandler returns all available cards sorted by strength.
// @Summary Get all cards
// @Description Retrieve all available hero and monster cards
// @Tags cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Card
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Router /cards [get]
func CardsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if the request context has been cancelled (shutdown in progress)
		select {
		case <-r.Context().Done():
			http.Error(w, "Server shutting down", http.StatusServiceUnavailable)
			return
		default:
		}

		cards, err := db.GetAllCards()
		if err != nil {
			http.Error(w, "Failed to fetch cards", http.StatusInternalServerError)
			return
		}

		// Sort cards by strength (strongest to weakest)
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].Strength > cards[j].Strength
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cards)
	}
}

// RegisterCardsRoutes registers the cards routes with the router.
func RegisterCardsRoutes(r chi.Router) {
	r.Get("/api/cards", CardsHandler())
}
