package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration values.
type Config struct {
	HTTPPort    string
	DatabaseURL string
	JWTSecret   string
}

// Load reads configuration from environment variables.
// Returns an error if required variables are missing.
func Load() (Config, error) {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return Config{
		HTTPPort:    port,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
	}, nil
}
