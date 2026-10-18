package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		slog.Error("catalog seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if os.Getenv("SEED_CONFIRM") != "true" {
		return errors.New("SEED_CONFIRM=true is required for explicit catalog loading")
	}

	kind := os.Getenv("SEED_KIND")
	if kind == "" {
		return errors.New("SEED_KIND is required (supported: reference)")
	}
	if kind != "reference" {
		return fmt.Errorf("unsupported seed kind %q: demo or development fixtures are strictly disabled in this repository", kind)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	dir, err := resolveCatalogDir(kind)
	if err != nil {
		return err
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

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read catalog directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		return fmt.Errorf("no SQL files found in %s", dir)
	}

	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute %s: %w", file, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", file, err)
		}

		slog.Info("applied reference catalog file", "file", file)
	}

	slog.Info("reference catalog successfully loaded", "kind", kind, "files_count", len(files))
	return nil
}

func resolveCatalogDir(kind string) (string, error) {
	candidates := []string{
		filepath.Join("catalogs", kind),
		filepath.Join("backend", "catalogs", kind),
		filepath.Join("..", "catalogs", kind),
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return filepath.Abs(candidate)
		}
	}
	return "", fmt.Errorf("catalog directory for kind %q not found", kind)
}
