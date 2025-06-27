package config

import "os"

// Config holds runtime configuration.
type Config struct {
	Port            string
	PocketBaseURL   string
	PocketBaseToken string
	PocketBaseAOKey string
}

// Load reads env vars and returns a Config.
func Load() Config {
	return Config{
		Port:            getenv("PORT", "8080"),
		PocketBaseURL:   getenv("POCKETBASE_URL", "http://localhost:8090"),
		PocketBaseToken: getenv("POCKETBASE_TOKEN", ""),
		PocketBaseAOKey: getenv("POCKETBASE_AO_KEY", ""),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
