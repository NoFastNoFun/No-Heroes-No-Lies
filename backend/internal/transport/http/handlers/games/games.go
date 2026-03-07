package games

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/models"
	"github.com/no-heroes-no-lies/backend/internal/repository"
	"github.com/no-heroes-no-lies/backend/internal/service"
	"github.com/no-heroes-no-lies/backend/internal/transport/http/middleware"
	"github.com/no-heroes-no-lies/backend/pkg/response"
)

type CreateRequest struct {
	MaxPlayers int `json:"max_players"`
}

func Create(lobby *service.LobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.JSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		playerID, ok := middleware.GetPlayerID(r.Context())
		if !ok {
			response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		req := CreateRequest{MaxPlayers: 4}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.MaxPlayers < 2 {
			req.MaxPlayers = 4
		}
		game, err := lobby.CreateGame(r.Context(), playerID, req.MaxPlayers)
		if err != nil {
			if err == service.ErrInvalidInput {
				response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid max_players"})
				return
			}
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		response.JSON(w, http.StatusCreated, game)
	}
}

func List(lobby *service.LobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			response.JSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		_, ok := middleware.GetPlayerID(r.Context())
		if !ok {
			response.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		list, err := lobby.ListGames(r.Context(), models.GameStatusLobby)
		if err != nil {
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		response.JSON(w, http.StatusOK, list)
	}
}

func Get(lobby *service.LobbyService, db *repository.DB) http.HandlerFunc {
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
		game, err := lobby.GetGame(r.Context(), gameID)
		if err != nil {
			response.JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		inGame, _ := db.IsPlayerInGame(r.Context(), gameID, playerID)
		out := map[string]interface{}{
			"id": game.ID, "creator_id": game.CreatorID, "status": game.Status,
			"max_players": game.MaxPlayers, "current_turn_index": game.CurrentTurnIndex,
			"round_number": game.RoundNumber, "created_at": game.CreatedAt,
			"in_game": inGame,
		}
		response.JSON(w, http.StatusOK, out)
	}
}

func Join(lobby *service.LobbyService, sessionService *service.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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
		slot, err := lobby.JoinGame(r.Context(), gameID, playerID)
		if err != nil {
			if err == service.ErrConflict {
				response.JSON(w, http.StatusConflict, map[string]string{"error": "game full or already joined"})
				return
			}
			if err == service.ErrBadRequest {
				response.JSON(w, http.StatusBadRequest, map[string]string{"error": "game not in lobby"})
				return
			}
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		_ = sessionService.ExtendSessionForPlayer(r.Context(), playerID)
		response.JSON(w, http.StatusOK, map[string]int{"slot_index": slot})
	}
}

func Start(lobby *service.LobbyService, sessionService *service.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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
		err = lobby.StartGame(r.Context(), gameID, playerID)
		if err != nil {
			if err == service.ErrForbidden {
				response.JSON(w, http.StatusForbidden, map[string]string{"error": "not game master"})
				return
			}
			if err == service.ErrBadRequest {
				response.JSON(w, http.StatusBadRequest, map[string]string{"error": "slots not filled or game not in lobby"})
				return
			}
			response.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
		players, _ := lobby.GetGamePlayers(r.Context(), gameID)
		for _, p := range players {
			_ = sessionService.ExtendSessionForPlayer(r.Context(), p.PlayerID)
		}
		response.JSON(w, http.StatusOK, map[string]string{"status": "started"})
	}
}
