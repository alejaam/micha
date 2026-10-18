package migrations

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationsTable = "schema_migrations"

// Apply executes SQL migrations from a directory in lexicographical order.
// Each file is executed once inside a transaction, tracked with SHA-256 checksum,
// and protected by a PostgreSQL advisory lock against concurrent runs.
func Apply(ctx context.Context, db *pgxpool.Pool, dir string) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("apply migrations: ensure table: %w", err)
	}
	if _, err := db.Exec(ctx, `SELECT pg_advisory_lock(hashtextextended($1, 0))`, migrationsTable); err != nil {
		return fmt.Errorf("apply migrations: acquire lock: %w", err)
	}
	defer func() {
		_, _ = db.Exec(ctx, `SELECT pg_advisory_unlock(hashtextextended($1, 0))`, migrationsTable)
	}()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("apply migrations: read dir %q: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		applied, checkErr := isApplied(ctx, db, file, fullPath)
		if checkErr != nil {
			return fmt.Errorf("apply migrations: check %s: %w", file, checkErr)
		}
		if applied {
			continue
		}
		if err := applyFile(ctx, db, fullPath, file); err != nil {
			return fmt.Errorf("apply migrations: %s: %w", file, err)
		}
	}
	return nil
}

func ensureMigrationsTable(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			checksum TEXT NOT NULL DEFAULT ''
		)
	`)
	return err
}

func isApplied(ctx context.Context, db *pgxpool.Pool, filename, fullPath string) (bool, error) {
	var checksum string
	err := db.QueryRow(ctx, `SELECT checksum FROM `+migrationsTable+` WHERE filename = $1`, filename).Scan(&checksum)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if checksum == "" {
		return true, nil
	}
	current, err := fileChecksum(fullPath)
	if err != nil {
		return false, err
	}
	if checksum != current {
		return false, fmt.Errorf("checksum mismatch for applied migration %s (db: %s, file: %s)", filename, checksum, current)
	}
	return true, nil
}

func applyFile(ctx context.Context, db *pgxpool.Pool, fullPath, filename string) error {
	sqlBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	checksum := sha256.Sum256(sqlBytes)

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("exec statement: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO `+migrationsTable+` (filename, checksum) VALUES ($1, $2)`,
		filename, fmt.Sprintf("%x", checksum[:]),
	); err != nil {
		return fmt.Errorf("insert migration row: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func fileChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file for checksum: %w", err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum[:]), nil
}
