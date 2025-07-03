package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret []byte

// SetJWTSecret sets the secret for signing JWTs.
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// SignJWT creates a JWT for the given userID.
func SignJWT(userID string, _ string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyJWT parses and validates the JWT, returning userID.
func VerifyJWT(tokenString string) (userID string, _ string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid, _ := claims["sub"].(string)
		return uid, "", nil
	}
	return "", "", jwt.ErrTokenMalformed
}
