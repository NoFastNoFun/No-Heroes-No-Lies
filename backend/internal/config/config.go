package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            int
	DatabaseURL     string
	JWTSecret       string
	SessionTTL      time.Duration
	ChallengeWindow time.Duration
}

func Load() (*Config, error) {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/nhml?sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}
	sessionTTL := 24 * time.Hour
	if s := os.Getenv("SESSION_TTL"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			sessionTTL = d
		}
	}
	challengeWindow := 10 * time.Second
	if s := os.Getenv("CHALLENGE_WINDOW"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			challengeWindow = d
		}
	}
	return &Config{
		Port:            port,
		DatabaseURL:     databaseURL,
		JWTSecret:       jwtSecret,
		SessionTTL:      sessionTTL,
		ChallengeWindow: challengeWindow,
	}, nil
}
