package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsTable = "schema_migrations"

// Apply executes SQL migrations from a directory in lexicographical order.
// Each file is applied once and recorded in schema_migrations.
func Apply(ctx context.Context, db *pgxpool.Pool, dir string) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("apply migrations: ensure table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("apply migrations: read dir %q: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	for _, file := range files {
		applied, checkErr := isApplied(ctx, db, file)
		if checkErr != nil {
			return fmt.Errorf("apply migrations: check %s: %w", file, checkErr)
		}
		if applied {
			continue
		}

		if err := applyFile(ctx, db, filepath.Join(dir, file), file); err != nil {
			return fmt.Errorf("apply migrations: %s: %w", file, err)
		}
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func isApplied(ctx context.Context, db *pgxpool.Pool, filename string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM `+migrationsTable+` WHERE filename = $1)`,
		filename,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func applyFile(ctx context.Context, db *pgxpool.Pool, fullPath, filename string) error {
	sqlBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, execErr := tx.Exec(ctx, string(sqlBytes)); execErr != nil {
		return fmt.Errorf("exec statement: %w", execErr)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO `+migrationsTable+` (filename) VALUES ($1)`,
		filename,
	); err != nil {
		return fmt.Errorf("insert migration row: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
