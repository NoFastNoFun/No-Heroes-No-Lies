package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/db"

	"golang.org/x/crypto/bcrypt"
)

var jwtExpiry = 15 * time.Minute // can be made configurable

// AuthRegisterHandler handles POST /auth/register
func AuthRegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Username == "" || req.Password == "" {
			http.Error(w, "Missing fields", http.StatusBadRequest)
			return
		}
		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		user, err := db.CreateUser(req.Email, req.Username, string(hash))
		if err != nil {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		jwtToken, err := auth.SignJWT(user.ID, "", jwtExpiry)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		setAuthCookie(w, jwtToken)
		w.WriteHeader(http.StatusNoContent)
	}
}

// AuthLoginHandler handles POST /auth/login
func AuthLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		user, err := db.GetUserByEmail(req.Email)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		jwtToken, err := auth.SignJWT(user.ID, "", jwtExpiry)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		setAuthCookie(w, jwtToken)
		w.WriteHeader(http.StatusNoContent)
	}
}

// AuthRefreshHandler handles POST /auth/refresh
func AuthRefreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwtToken, err := readAuthCookie(r)
		if err != nil {
			http.Error(w, "Missing auth cookie", http.StatusUnauthorized)
			return
		}
		jwtToken, err = auth.SignJWT("", "", jwtExpiry)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		setAuthCookie(w, jwtToken)
		w.WriteHeader(http.StatusNoContent)
	}
}

// AuthLogoutHandler handles GET /auth/logout
func AuthLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearAuthCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

// setAuthCookie sets the game_auth cookie.
func setAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "game_auth",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(jwtExpiry.Seconds()),
	}
	http.SetCookie(w, cookie)
}

// clearAuthCookie clears the game_auth cookie.
func clearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "game_auth",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

// readAuthCookie reads the game_auth cookie from the request.
func readAuthCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("game_auth")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
