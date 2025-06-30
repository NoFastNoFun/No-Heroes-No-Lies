package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret []byte

// SetJWTSecret sets the secret for signing JWTs.
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// SignJWT creates a JWT for the given userID and PB access token hash.
func SignJWT(userID, pbAccessToken string, expiry time.Duration) (string, error) {
	hash := sha256.Sum256([]byte(pbAccessToken))
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(expiry).Unix(),
		"pb":  hex.EncodeToString(hash[:]),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyJWT parses and validates the JWT, returning userID and pb token hash.
func VerifyJWT(tokenString string) (userID, pbHash string, err error) {
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
		pb, _ := claims["pb"].(string)
		return uid, pb, nil
	}
	return "", "", jwt.ErrTokenMalformed
}
