package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration values.
type Config struct {
	HTTPPort           string
	DatabaseURL        string
	JWTSecret          string
	AllowedOrigins     []string
	AllowOwnerOnBehalf bool
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
		HTTPPort:           port,
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		AllowedOrigins:     parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
		AllowOwnerOnBehalf: parseBoolDefaultTrue(os.Getenv("ALLOW_OWNER_ON_BEHALF")),
	}, nil
}

// parseAllowedOrigins splits a comma-separated list of origins.
// Returns ["*"] if the input is empty (allow all in development).
func parseAllowedOrigins(raw string) []string {
	if raw == "" {
		return []string{"*"}
	}
	origins := strings.Split(raw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
}

func parseBoolDefaultTrue(raw string) bool {
	if strings.TrimSpace(raw) == "" {
		return true
	}
	v, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return true
	}
	return v
}
