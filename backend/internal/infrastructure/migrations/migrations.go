package migrations

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	migrationsTable        = "schema_migrations"
	baselineFileName       = "001_initial_schema.sql"
	initialSchemaChecksum  = "b2d77e9dda2ee7d45297eb001bc042bba373c63aa76b80acaa19a030ea9317b8"
)

// Apply executes SQL migrations from a directory in lexicographical order.
// Each file is executed once inside a transaction, tracked with SHA-256 checksum,
// and protected by a PostgreSQL advisory lock against concurrent runs.
// All operations run on a single acquired connection to avoid deadlocks when MaxConns=1.
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
		// Ensure advisory lock is released even if caller context was cancelled
		unlockCtx := context.WithoutCancel(ctx)
		if _, unlockErr := conn.Exec(unlockCtx, `SELECT pg_advisory_unlock(hashtextextended($1, 0))`, migrationsTable); unlockErr != nil {
			slog.Warn("apply migrations: failed to release advisory lock", "error", unlockErr)
		}
	}()

	if err := ensureMigrationsTable(ctx, conn); err != nil {
		return fmt.Errorf("apply migrations: ensure table: %w", err)
	}

	if err := reconcileBaseline(ctx, conn, dir); err != nil {
		return fmt.Errorf("apply migrations: baseline: %w", err)
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
		applied, checkErr := isApplied(ctx, conn, file, fullPath)
		if checkErr != nil {
			return fmt.Errorf("apply migrations: check %s: %w", file, checkErr)
		}
		if applied {
			continue
		}
		if applyErr := applyFile(ctx, conn, fullPath, file); applyErr != nil {
			return fmt.Errorf("apply migrations: %s: %w", file, applyErr)
		}
	}
	return nil
}

func ensureMigrationsTable(ctx context.Context, conn *pgxpool.Conn) error {
	if _, createErr := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			checksum TEXT NOT NULL DEFAULT ''
		)
	`); createErr != nil {
		return createErr
	}
	_, alterErr := conn.Exec(ctx, `ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS checksum TEXT NOT NULL DEFAULT ''`)
	return alterErr
}

// reconcileBaseline checks if the database already ran legacy pre-v1 migrations (001-033)
// before the baseline consolidation. If fully migrated to 033, it reconciles the baseline.
// If partially migrated, it aborts startup with an explicit database recreation requirement.
func reconcileBaseline(ctx context.Context, conn *pgxpool.Conn, dir string) error {
	baselinePath := filepath.Join(dir, baselineFileName)
	if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
		return nil
	}

	var dummy string
	err := conn.QueryRow(ctx, `SELECT filename FROM `+migrationsTable+` WHERE filename = $1`, baselineFileName).Scan(&dummy)
	if err == nil {
		return nil // baseline already recorded
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("check baseline in migrations table: %w", err)
	}

	var legacyCount int
	if countErr := conn.QueryRow(ctx, `SELECT COUNT(*) FROM `+migrationsTable).Scan(&legacyCount); countErr != nil {
		return fmt.Errorf("check legacy migrations count: %w", countErr)
	}

	if legacyCount == 0 {
		return nil // clean database, normal flow will apply baseline
	}

	var legacy33File string
	err = conn.QueryRow(ctx, `SELECT filename FROM `+migrationsTable+` WHERE filename LIKE '%033%' OR filename = '033_add_name_to_users.sql' LIMIT 1`).Scan(&legacy33File)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("database contains %d legacy migrations but is missing the final legacy migration (033_add_name_to_users.sql); schema differences cannot be safely resolved. Please recreate the database (e.g. docker compose down -v)", legacyCount)
	}
	if err != nil {
		return fmt.Errorf("check legacy 033 migration: %w", err)
	}

	slog.Warn("reconciling baseline migration against existing legacy database schema", "baseline", baselineFileName, "legacy_migrations_count", legacyCount)
	if _, insertErr := conn.Exec(ctx, `INSERT INTO `+migrationsTable+` (filename, checksum) VALUES ($1, $2)`, baselineFileName, initialSchemaChecksum); insertErr != nil {
		return fmt.Errorf("reconcile baseline %s: %w", baselineFileName, insertErr)
	}
	return nil
}

func isApplied(ctx context.Context, conn *pgxpool.Conn, filename, fullPath string) (bool, error) {
	var checksum string
	err := conn.QueryRow(ctx, `SELECT checksum FROM `+migrationsTable+` WHERE filename = $1`, filename).Scan(&checksum)
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
		// Only safely backfill if matching the trusted canonical baseline checksum
		if filename == baselineFileName && current == initialSchemaChecksum {
			if _, updateErr := conn.Exec(ctx, `UPDATE `+migrationsTable+` SET checksum = $1 WHERE filename = $2`, initialSchemaChecksum, filename); updateErr != nil {
				return false, fmt.Errorf("backfill checksum for %s: %w", filename, updateErr)
			}
			return true, nil
		}
		return false, fmt.Errorf("empty checksum for applied migration %s: cannot verify against disk without trusted checksum", filename)
	}

	if checksum != current {
		return false, fmt.Errorf("checksum mismatch for applied migration %s (db: %s, file: %s)", filename, checksum, current)
	}
	return true, nil
}

func applyFile(ctx context.Context, conn *pgxpool.Conn, fullPath, filename string) error {
	sqlBytes, readErr := os.ReadFile(fullPath)
	if readErr != nil {
		return fmt.Errorf("read file: %w", readErr)
	}
	checksum := sha256.Sum256(sqlBytes)

	tx, txErr := conn.Begin(ctx)
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
