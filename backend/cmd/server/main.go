package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/no-heroes-no-lies/backend/internal/config"
	"github.com/no-heroes-no-lies/backend/internal/migrations"
	"github.com/no-heroes-no-lies/backend/internal/repository"
	"github.com/no-heroes-no-lies/backend/internal/transport/http/router"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := repository.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := repository.RunMigrations(ctx, db.Pool, migrations.FS, "."); err != nil {
		slog.Error("migrations", "err", err)
		os.Exit(1)
	}

	r := router.New(cfg, db)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
	slog.Info("server stopped")
}
