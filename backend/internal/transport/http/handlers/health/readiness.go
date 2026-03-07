package health

import (
	"context"
	"net/http"

	"github.com/no-heroes-no-lies/backend/internal/repository"
)

func Readiness(db *repository.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		if err := db.Pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("db unavailable"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}
