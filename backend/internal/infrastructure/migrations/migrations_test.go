package migrations

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestFileChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.sql")

	content := []byte("CREATE TABLE test (id INT);")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	sum1, err := fileChecksum(filePath)
	if err != nil {
		t.Fatalf("fileChecksum failed: %v", err)
	}
	if sum1 == "" {
		t.Error("expected non-empty checksum")
	}

	sum2, err := fileChecksum(filePath)
	if err != nil {
		t.Fatalf("fileChecksum second run failed: %v", err)
	}
	if sum1 != sum2 {
		t.Errorf("expected deterministic checksum, got %s and %s", sum1, sum2)
	}

	// Modifying content should change checksum
	if err := os.WriteFile(filePath, []byte("CREATE TABLE test (id INT, name TEXT);"), 0644); err != nil {
		t.Fatalf("failed to modify test file: %v", err)
	}
	sum3, err := fileChecksum(filePath)
	if err != nil {
		t.Fatalf("fileChecksum modified run failed: %v", err)
	}
	if sum1 == sum3 {
		t.Error("expected checksum to change when content changes")
	}
}

func TestApply_Integration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("skipping migration integration test: DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("skipping: database unreachable: %v", err)
	}

	tmpDir := t.TempDir()
	mig1 := filepath.Join(tmpDir, "001_test_migration.sql")
	if err := os.WriteFile(mig1, []byte("CREATE TABLE IF NOT EXISTS _test_table (id TEXT PRIMARY KEY);"), 0644); err != nil {
		t.Fatalf("failed to write mig1: %v", err)
	}

	if err := Apply(ctx, pool, tmpDir); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Clean up test table and test migration entry
	defer func() {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS _test_table;")
		_, _ = pool.Exec(ctx, "DELETE FROM schema_migrations WHERE filename = '001_test_migration.sql';")
	}()

	// Re-applying without changes should succeed (idempotent)
	if err := Apply(ctx, pool, tmpDir); err != nil {
		t.Fatalf("second Apply should succeed idempotently: %v", err)
	}

	// Test backfill: clear checksum in DB and re-apply
	if _, err := pool.Exec(ctx, "UPDATE schema_migrations SET checksum = '' WHERE filename = '001_test_migration.sql'"); err != nil {
		t.Fatalf("clear checksum failed: %v", err)
	}
	if err := Apply(ctx, pool, tmpDir); err != nil {
		t.Fatalf("Apply should backfill empty checksum: %v", err)
	}
	var backfilledChecksum string
	_ = pool.QueryRow(ctx, "SELECT checksum FROM schema_migrations WHERE filename = '001_test_migration.sql'").Scan(&backfilledChecksum)
	if backfilledChecksum == "" {
		t.Error("expected checksum to be backfilled, got empty")
	}

	// Tampering with the file should trigger checksum error
	if err := os.WriteFile(mig1, []byte("CREATE TABLE IF NOT EXISTS _test_table (id TEXT PRIMARY KEY, extra TEXT);"), 0644); err != nil {
		t.Fatalf("failed to tamper mig1: %v", err)
	}

	err = Apply(ctx, pool, tmpDir)
	if err == nil {
		t.Fatal("expected checksum mismatch error after tampering applied migration")
	}
}
