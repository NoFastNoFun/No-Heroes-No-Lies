package config

import "os"

// Config holds runtime configuration.
type Config struct {
	Port                string
	AllowedDomainSuffix string
	RedisAddr           string
	GameJWTSecret       string
}

// Load reads env vars and returns a Config.
func Load() Config {
	return Config{
		Port:                getenv("PORT", "8080"),
		AllowedDomainSuffix: getenv("ALLOWED_DOMAIN_SUFFIX", "localhost:8080"),
		RedisAddr:           getenv("REDIS_ADDR", "localhost:6379"),
		GameJWTSecret:       getenv("GAME_JWT_SECRET", "changeme"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
