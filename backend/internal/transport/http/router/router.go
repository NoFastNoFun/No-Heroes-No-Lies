package router

import (
	"net/http"

	"github.com/no-heroes-no-lies/backend/internal/config"
	"github.com/no-heroes-no-lies/backend/internal/gameloop"
	"github.com/no-heroes-no-lies/backend/internal/powers"
	"github.com/no-heroes-no-lies/backend/internal/repository"
	"github.com/no-heroes-no-lies/backend/internal/service"
	"github.com/no-heroes-no-lies/backend/internal/transport/http/handlers/games"
	"github.com/no-heroes-no-lies/backend/internal/transport/http/handlers/health"
	"github.com/no-heroes-no-lies/backend/internal/transport/http/handlers/session"
	"github.com/no-heroes-no-lies/backend/internal/transport/http/middleware"
	ws "github.com/no-heroes-no-lies/backend/internal/transport/http/ws"
)

func New(cfg *config.Config, db *repository.DB) http.Handler {
	mux := http.NewServeMux()

	sessionService := service.NewSessionService(db, cfg)
	powerReg := powers.NewRegistry()
	powers.RegisterAll(powerReg)
	wsHub := ws.NewHub(nil, sessionService, db)
	gameNotifier := ws.NewNotifier(wsHub)
	turnProcessor := service.NewTurnProcessor(db, cfg, powerReg, gameNotifier)
	gameRegistry := gameloop.NewRegistry(turnProcessor)
	wsHub.SetRegistry(gameRegistry)
	lobbyService := service.NewLobbyService(db, gameRegistry)
	stateService := service.NewGameStateService(db)

	mux.HandleFunc("GET /health", health.Liveness())
	mux.HandleFunc("GET /ready", health.Readiness(db))
	mux.Handle("POST /api/v1/session", session.Create(sessionService))

	auth := middleware.Auth(sessionService)
	mux.Handle("POST /api/v1/games", auth(games.Create(lobbyService)))
	mux.Handle("GET /api/v1/games", auth(games.List(lobbyService)))
	mux.Handle("GET /api/v1/games/{id}", auth(games.Get(lobbyService, db)))
	mux.Handle("GET /api/v1/games/{id}/state", auth(games.State(stateService)))
	mux.Handle("POST /api/v1/games/{id}/join", auth(games.Join(lobbyService, sessionService)))
	mux.Handle("POST /api/v1/games/{id}/start", auth(games.Start(lobbyService, sessionService)))
	mux.Handle("GET /api/v1/games/{id}/ws", ws.Handler(wsHub))

	return middleware.WithLogging(mux)
}
