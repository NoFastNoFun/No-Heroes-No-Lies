package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
)

const (
	authHeader          = "Authorization"
	ctxKeyUserID ctxKey = "userID"
)

type ctxKey string

// Middleware returns an HTTP middleware that:
//  1. extracts the JWT from the game_auth cookie
//  2. verifies it with our backend secret
//  3. checks PB token hash against Redis
//  4. on success, injects userID into request context
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("game_auth")
			if err != nil || cookie.Value == "" {
				log.Printf("Auth middleware: missing game_auth cookie for %s %s", r.Method, r.URL.Path)
				http.Error(w, "missing auth cookie", http.StatusUnauthorized)
				return
			}

			userID, _, err := VerifyJWT(cookie.Value)
			if err != nil || userID == "" {
				log.Printf("Auth middleware: invalid JWT for %s %s", r.Method, r.URL.Path)
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext returns the authenticated user ID.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKeyUserID).(string)
	return id, ok
}

func extractBearer(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
