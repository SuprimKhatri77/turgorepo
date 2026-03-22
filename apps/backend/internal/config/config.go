package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration from environment.
type Config struct {
	Port        int
	GinMode     string
	DatabaseURL string
}

// Load loads .env from the current directory (if present) then reads configuration from environment variables.
// No hardcoded URLs: PORT defaults to 5000 if unset; GIN_MODE defaults to "debug"; DATABASE_URL must be set in .env or env.
func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error if .env is missing

	port := 5000
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	dbURL := os.Getenv("DATABASE_URL")

	return &Config{
		Port:        port,
		GinMode:     ginMode,
		DatabaseURL: dbURL,
	}, nil
}
