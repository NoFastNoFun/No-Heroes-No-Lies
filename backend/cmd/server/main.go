package main

import (
	"log"
	"net/http"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/config"
	"no-heroes-no-lies/internal/handlers"
	"no-heroes-no-lies/internal/hostfilter"
	"no-heroes-no-lies/internal/pb"
	"no-heroes-no-lies/internal/services"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	pbClient := pb.NewClient(cfg.PocketBaseURL, cfg.PocketBaseAOKey)
	gameService := services.NewGameService(pbClient)

	r := chi.NewRouter()
	r.Use(auth.Middleware(pbClient))
	r.Use(hostfilter.Middleware(cfg.AllowedDomainSuffix))
	r.Use(auth.Middleware(pbClient))

	// liveness probe
	r.Get("/api/health", handlers.HealthHandler)
	handlers.RegisterGameRoutes(r, gameService)

	addr := ":" + cfg.Port
	log.Printf("server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
