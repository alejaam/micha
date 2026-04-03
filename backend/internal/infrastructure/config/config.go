package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all runtime configuration values.
type Config struct {
	HTTPPort       string
	DatabaseURL    string
	JWTSecret      string
	AllowedOrigins []string
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
	if len(strings.TrimSpace(jwtSecret)) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	allowedOrigins, err := parseAllowedOrigins(
		os.Getenv("ALLOWED_ORIGINS"),
		isProductionEnv(os.Getenv("APP_ENV"), os.Getenv("ENV")),
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPPort:       port,
		DatabaseURL:    dbURL,
		JWTSecret:      jwtSecret,
		AllowedOrigins: allowedOrigins,
	}, nil
}

// parseAllowedOrigins splits a comma-separated list of origins.
// Returns ["*"] when running outside production and no value is provided.
func parseAllowedOrigins(raw string, production bool) ([]string, error) {
	if raw == "" {
		if production {
			return nil, fmt.Errorf("ALLOWED_ORIGINS environment variable is required in production")
		}
		return []string{"*"}, nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, origin := range parts {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if production && trimmed == "*" {
			return nil, fmt.Errorf("ALLOWED_ORIGINS cannot contain wildcard '*' in production")
		}
		origins = append(origins, trimmed)
	}

	if len(origins) == 0 {
		if production {
			return nil, fmt.Errorf("ALLOWED_ORIGINS must contain at least one origin in production")
		}
		return []string{"*"}, nil
	}

	return origins, nil
}

func isProductionEnv(appEnv, env string) bool {
	active := strings.ToLower(strings.TrimSpace(appEnv))
	if active == "" {
		active = strings.ToLower(strings.TrimSpace(env))
	}
	return active == "prod" || active == "production"
}
