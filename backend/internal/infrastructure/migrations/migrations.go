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
	conn, connErr := db.Acquire(ctx)
	if connErr != nil {
		return fmt.Errorf("apply migrations: acquire pool connection: %w", connErr)
	}
	defer conn.Release()

	if _, lockErr := conn.Exec(ctx, `SELECT pg_advisory_lock(hashtextextended($1, 0))`, migrationsTable); lockErr != nil {
		return fmt.Errorf("apply migrations: acquire lock: %w", lockErr)
	}
	defer func() {
		_, _ = conn.Exec(ctx, `SELECT pg_advisory_unlock(hashtextextended($1, 0))`, migrationsTable)
	}()

	if err := ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("apply migrations: ensure table: %w", err)
	}

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
		if applyErr := applyFile(ctx, db, fullPath, file); applyErr != nil {
			return fmt.Errorf("apply migrations: %s: %w", file, applyErr)
		}
	}
	return nil
}

func ensureMigrationsTable(ctx context.Context, db *pgxpool.Pool) error {
	if _, createErr := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			checksum TEXT NOT NULL DEFAULT ''
		)
	`); createErr != nil {
		return createErr
	}
	_, alterErr := db.Exec(ctx, `ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS checksum TEXT NOT NULL DEFAULT ''`)
	return alterErr
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

	current, readErr := fileChecksum(fullPath)
	if readErr != nil {
		return false, readErr
	}

	if checksum == "" {
		// Backfill missing checksum for historical migration so future changes are detected
		if _, updateErr := db.Exec(ctx, `UPDATE `+migrationsTable+` SET checksum = $1 WHERE filename = $2`, current, filename); updateErr != nil {
			return false, fmt.Errorf("backfill checksum for %s: %w", filename, updateErr)
		}
		return true, nil
	}

	if checksum != current {
		return false, fmt.Errorf("checksum mismatch for applied migration %s (db: %s, file: %s)", filename, checksum, current)
	}
	return true, nil
}

func applyFile(ctx context.Context, db *pgxpool.Pool, fullPath, filename string) error {
	sqlBytes, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		return fmt.Errorf("read file: %w", readErr)
	}
	checksum := sha256.Sum256(sqlBytes)

	tx, txErr := db.Begin(ctx)
	if txErr != nil {
		return fmt.Errorf("begin tx: %w", txErr)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, execErr := tx.Exec(ctx, string(sqlBytes)); execErr != nil {
		return fmt.Errorf("exec statement: %w", execErr)
	}
	if _, insertErr := tx.Exec(ctx,
		`INSERT INTO `+migrationsTable+` (filename, checksum) VALUES ($1, $2)`,
		filename, fmt.Sprintf("%x", checksum[:]),
	); insertErr != nil {
		return fmt.Errorf("insert migration row: %w", insertErr)
	}
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("commit tx: %w", commitErr)
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
