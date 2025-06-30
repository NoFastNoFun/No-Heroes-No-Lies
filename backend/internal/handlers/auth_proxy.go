package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"no-heroes-no-lies/internal/auth"
	"no-heroes-no-lies/internal/pb"
	"no-heroes-no-lies/internal/services"
)

var jwtExpiry = 15 * time.Minute // can be made configurable

// AuthLoginHandler handles POST /auth/login
func AuthLoginHandler(pbClient *pb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		userID, accessToken, refreshToken, _, err := pbClient.AuthWithPasswordFull(req.Email, req.Password)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		sess := services.PBSession{AccessToken: accessToken, RefreshToken: refreshToken}
		err = services.SetPBSession(r.Context(), userID, sess, jwtExpiry)
		if err != nil {
			log.Printf("Failed to store PB session: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		jwtToken, err := auth.SignJWT(userID, accessToken, jwtExpiry)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		setAuthCookie(w, jwtToken)
		w.WriteHeader(http.StatusNoContent)
	}
}

// AuthRefreshHandler handles POST /auth/refresh
func AuthRefreshHandler(pbClient *pb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwtToken, err := readAuthCookie(r)
		if err != nil {
			http.Error(w, "Missing auth cookie", http.StatusUnauthorized)
			return
		}
		userID, _, err := auth.VerifyJWT(jwtToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		sess, err := services.GetPBSession(r.Context(), userID)
		if err != nil {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}
		newAccessToken, newRefreshToken, err := pbClient.AuthRefresh(sess.AccessToken, sess.RefreshToken)
		if err != nil {
			http.Error(w, "Failed to refresh", http.StatusUnauthorized)
			return
		}
		sess = services.PBSession{AccessToken: newAccessToken, RefreshToken: newRefreshToken}
		err = services.SetPBSession(r.Context(), userID, sess, jwtExpiry)
		if err != nil {
			log.Printf("Failed to update PB session: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		jwtToken, err = auth.SignJWT(userID, newAccessToken, jwtExpiry)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		setAuthCookie(w, jwtToken)
		w.WriteHeader(http.StatusNoContent)
	}
}

// AuthLogoutHandler handles GET /auth/logout
func AuthLogoutHandler(pbClient *pb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwtToken, err := readAuthCookie(r)
		if err != nil {
			clearAuthCookie(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		userID, _, err := auth.VerifyJWT(jwtToken)
		if err != nil {
			clearAuthCookie(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		sess, err := services.GetPBSession(r.Context(), userID)
		if err == nil {
			_ = pbClient.AuthLogout(sess.AccessToken) // ignore error
		}
		_ = services.DeletePBSession(r.Context(), userID)
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
