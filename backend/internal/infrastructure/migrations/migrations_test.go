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
	if writeErr := os.WriteFile(filePath, content, 0644); writeErr != nil {
		t.Fatalf("failed to write test file: %v", writeErr)
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
	if writeErr := os.WriteFile(filePath, []byte("CREATE TABLE test (id INT, name TEXT);"), 0644); writeErr != nil {
		t.Fatalf("failed to modify test file: %v", writeErr)
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

	if pingErr := pool.Ping(ctx); pingErr != nil {
		t.Skipf("skipping: database unreachable: %v", pingErr)
	}

	tmpDir := t.TempDir()
	mig1 := filepath.Join(tmpDir, "001_test_migration.sql")
	if writeErr := os.WriteFile(mig1, []byte("CREATE TABLE IF NOT EXISTS _test_table (id TEXT PRIMARY KEY);"), 0644); writeErr != nil {
		t.Fatalf("failed to write mig1: %v", writeErr)
	}

	if applyErr := Apply(ctx, pool, tmpDir); applyErr != nil {
		t.Fatalf("Apply failed: %v", applyErr)
	}

	// Clean up test table and test migration entry
	defer func() {
		_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS _test_table;")
		_, _ = pool.Exec(ctx, "DELETE FROM schema_migrations WHERE filename = '001_test_migration.sql';")
	}()

	// Re-applying without changes should succeed (idempotent)
	if applyErr := Apply(ctx, pool, tmpDir); applyErr != nil {
		t.Fatalf("second Apply should succeed idempotently: %v", applyErr)
	}

	// Test empty checksum: clearing checksum on non-baseline migration must fail tamper check
	if _, execErr := pool.Exec(ctx, "UPDATE schema_migrations SET checksum = '' WHERE filename = '001_test_migration.sql'"); execErr != nil {
		t.Fatalf("clear checksum failed: %v", execErr)
	}
	if applyErr := Apply(ctx, pool, tmpDir); applyErr == nil {
		t.Fatalf("Apply should reject empty checksum on unverified migration, got nil")
	}

	// Restore valid checksum
	sum, err := fileChecksum(mig1)
	if err != nil {
		t.Fatalf("fileChecksum: %v", err)
	}
	if _, execErr := pool.Exec(ctx, "UPDATE schema_migrations SET checksum = $1 WHERE filename = '001_test_migration.sql'", sum); execErr != nil {
		t.Fatalf("restore checksum failed: %v", execErr)
	}

	// Tampering with the file should trigger checksum error
	if writeErr := os.WriteFile(mig1, []byte("CREATE TABLE IF NOT EXISTS _test_table (id TEXT PRIMARY KEY, extra TEXT);"), 0644); writeErr != nil {
		t.Fatalf("failed to tamper mig1: %v", writeErr)
	}

	err = Apply(ctx, pool, tmpDir)
	if err == nil {
		t.Fatal("expected checksum mismatch error after tampering applied migration")
	}
}
