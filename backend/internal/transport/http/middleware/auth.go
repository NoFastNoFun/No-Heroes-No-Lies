package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/no-heroes-no-lies/backend/internal/service"
)

type contextKey string

const PlayerIDKey contextKey = "player_id"

func Auth(sessionService *service.SessionService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing or invalid authorization")
				return
			}
			playerID, err := sessionService.ValidateToken(r.Context(), token)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			ctx := context.WithValue(r.Context(), PlayerIDKey, playerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}

func GetPlayerID(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(PlayerIDKey)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + escapeJSONString(msg) + `"}`))
}

func escapeJSONString(s string) string {
	var b []byte
	for _, r := range s {
		switch r {
		case '"', '\\':
			b = append(b, '\\', byte(r))
		default:
			b = append(b, byte(r))
		}
	}
	return string(b)
}
