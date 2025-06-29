package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"no-heroes-no-lies/internal/pb"

	"github.com/golang-jwt/jwt/v4"
)

type AuthRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}

var jwtSecret = []byte("supersecretkey_change_me") // TODO: move to config

func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// POST /api/auth/login
func LoginHandler(pbClient *pb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		// Authenticate with PocketBase
		resp, err := pbClient.AuthWithPassword(req.Email, req.Password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		jwtToken, err := generateJWT(resp.UserID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(AuthResponse{Token: jwtToken, User: resp.User})
	}
}

// POST /api/auth/register
func RegisterHandler(pbClient *pb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		// Register with PocketBase
		resp, err := pbClient.RegisterUser(req.Email, req.Password, req.Username, req.DisplayName)
		if err != nil {
			http.Error(w, "Registration failed", http.StatusBadRequest)
			return
		}
		jwtToken, err := generateJWT(resp.UserID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(AuthResponse{Token: jwtToken, User: resp.User})
	}
}
