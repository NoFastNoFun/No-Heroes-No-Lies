package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"strings"

	"no-heroes-no-lies/internal/services"
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

			userID, pbHash, err := VerifyJWT(cookie.Value)
			if err != nil {
				log.Printf("Auth middleware: JWT verification failed for %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			sess, err := services.GetPBSession(r.Context(), userID)
			if err != nil {
				log.Printf("Auth middleware: failed to load PB session for user %s: %v", userID, err)
				http.Error(w, "session expired", http.StatusUnauthorized)
				return
			}
			// Check PB access token hash
			hash := sha256.Sum256([]byte(sess.AccessToken))
			if hex.EncodeToString(hash[:]) != pbHash {
				log.Printf("Auth middleware: PB token hash mismatch for user %s", userID)
				http.Error(w, "token revoked", http.StatusUnauthorized)
				return
			}

			log.Printf("Auth middleware: JWT and PB token verified for user %s on %s %s", userID, r.Method, r.URL.Path)
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
