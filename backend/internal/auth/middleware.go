package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

const (
	authHeader            = "Authorization"
	ctxKeyPlayerID ctxKey = "playerID"
)

type ctxKey string

var jwtSecret = []byte("supersecretkey_change_me") // TODO: move to config

// Middleware returns an HTTP middleware that:
//  1. extracts the Bearer token (JWT)
//  2. verifies it with our backend secret
//  3. on success, injects playerID into request context
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearer(r.Header.Get(authHeader))
			if token == "" {
				log.Printf("Auth middleware: missing bearer token for %s %s", r.Method, r.URL.Path)
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			log.Printf("Auth middleware: verifying JWT for %s %s", r.Method, r.URL.Path)
			playerID, err := verifyJWT(token)
			if err != nil {
				log.Printf("Auth middleware: JWT verification failed for %s %s: %v", r.Method, r.URL.Path, err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			log.Printf("Auth middleware: JWT verified successfully for player %s on %s %s", playerID, r.Method, r.URL.Path)
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

func verifyJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["sub"].(string); ok {
			return userID, nil
		}
	}

	return "", errors.New("invalid token claims")
}
