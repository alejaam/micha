package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/infrastructure/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	environment := os.Getenv("SEED_ENVIRONMENT")
	scope := os.Getenv("SEED_SCOPE")
	kind := os.Getenv("SEED_KIND")
	if os.Getenv("SEED_CONFIRM") != "true" {
		return errors.New("SEED_CONFIRM=true is required for explicit seed execution")
	}
	if environment == "" || scope == "" || kind == "" {
		return errors.New("SEED_ENVIRONMENT, SEED_SCOPE, and SEED_KIND are required")
	}
	if environment != "development" && environment != "local" && environment != "staging" && environment != "production" && environment != "prod" {
		return fmt.Errorf("unsupported seed environment %q", environment)
	}
	switch kind {
	case "reference":
		if environment == "prod" {
			environment = "production"
		}
	case "development":
		if environment == "production" || environment == "prod" {
			return errors.New("development fixtures are disabled in production")
		}
		if os.Getenv("SEED_EMAIL") == "" {
			return errors.New("SEED_EMAIL is required for development seeds")
		}
	default:
		return fmt.Errorf("unsupported seed kind %q", kind)
	}

	dir, err := resolveSeedDir(kind)
	if err != nil {
		return err
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	settings := map[string]string{"app.seed_scope": scope}
	if email := os.Getenv("SEED_EMAIL"); email != "" {
		settings["app.seed_email"] = email
	}
	if err := migrations.ApplyWithOptions(ctx, pool, dir, migrations.Options{Settings: settings}); err != nil {
		return err
	}
	slog.Info("seed applied", "environment", environment, "scope", scope, "kind", kind)
	return nil
}

func resolveSeedDir(kind string) (string, error) {
	candidates := []string{
		filepath.Join("seeds", kind),
		filepath.Join("backend", "seeds", kind),
		filepath.Join("..", "seeds", kind),
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf("seed directory for kind %q not found", kind)
}
