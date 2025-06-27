package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"no-heroes-no-lies/internal/pb"
)

const (
	authHeader            = "Authorization"
	ctxKeyPlayerID ctxKey = "playerID"
)

type ctxKey string

// Middleware returns an HTTP middleware that:
//  1. extracts the Bearer token
//  2. verifies it with PocketBase
//  3. on success, injects playerID into request context
func Middleware(pbClient *pb.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r.Header.Get(authHeader))
			if token == "" {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			playerID, err := pbClient.VerifyUserToken(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyPlayerID, playerID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PlayerIDFromContext returns the authenticated player ID.
func PlayerIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value(ctxKeyPlayerID).(string)
	if !ok || id == "" {
		return "", errors.New("unauthenticated request")
	}
	return id, nil
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
