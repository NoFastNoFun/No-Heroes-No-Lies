// @title No Heroes No Lies API
// @version 1.0
// @description Game session management and gameplay API
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/config"
	"no-heroes-no-lies/internal/handlers"
	"no-heroes-no-lies/internal/hostfilter"
	"no-heroes-no-lies/internal/pb"
	"no-heroes-no-lies/internal/services"
	"no-heroes-no-lies/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	cfg := config.Load()

	// Set JWT secret
	auth.SetJWTSecret(cfg.GameJWTSecret)
	// Initialize Redis
	services.InitRedis(cfg.RedisAddr)

	pbClient := pb.NewClient(cfg.PocketBaseURL, cfg.PocketBaseAOKey)
	gameSvc := services.NewGameService(pbClient)
	sessionSvc := services.NewSessionService(pbClient)

	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(hostfilter.Middleware(cfg.AllowedDomainSuffix))

	r.Get("/api/health", handlers.HealthHandler)
	r.Get("/api/health/slow", handlers.SlowHealthHandler)

	// Auth routes (no auth required)
	r.Post("/api/auth/login", handlers.AuthLoginHandler(pbClient))
	r.Post("/api/auth/refresh", handlers.AuthRefreshHandler(pbClient))
	r.Get("/api/auth/logout", handlers.AuthLogoutHandler(pbClient))

	// OpenAPI documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+cfg.Port+"/docs/swagger.json"),
	))
	r.Get("/docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.json")
	})

	// Protected routes (require JWT auth)
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware())
		handlers.RegisterSessionRoutes(r, sessionSvc)
		handlers.RegisterGameRoutes(r, gameSvc)
		handlers.RegisterChallengeRoute(r, gameSvc)
		handlers.RegisterCardsRoutes(r, pbClient)
	})

	// WebSocket endpoint for real-time game updates
	r.Get("/ws/game/{id}", func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")
		ws.GameWSHandler(w, r, sessionID, gameSvc, sessionSvc)
	})

	// Create HTTP server
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// Kill (no param) default sends syscall.SIGTERM
	// Kill -2 is syscall.SIGINT
	// Kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// The context is used to inform the server it has 10 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server exited")
}
