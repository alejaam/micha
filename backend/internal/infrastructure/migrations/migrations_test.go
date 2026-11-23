package migrations

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestValidateBaselineSchema_Integration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("skipping schema validation integration test: DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	conn, acquireErr := pool.Acquire(ctx)
	if acquireErr != nil {
		t.Fatalf("acquire conn: %v", acquireErr)
	}
	defer conn.Release()

	// 1. Valid baseline schema check
	if valErr := validateBaselineSchema(ctx, conn); valErr != nil {
		t.Fatalf("expected valid baseline schema to pass validation: %v", valErr)
	}

	// 2. Extra table detection (e.g. obsolete table or rogue table)
	if _, createErr := conn.Exec(ctx, "CREATE TABLE public.period_approvals (id TEXT PRIMARY KEY)"); createErr != nil {
		t.Fatalf("create extra table: %v", createErr)
	}
	valErr := validateBaselineSchema(ctx, conn)
	if valErr == nil || !strings.Contains(valErr.Error(), `unexpected extra table "period_approvals"`) {
		t.Fatalf("expected unexpected extra table error for period_approvals, got: %v", valErr)
	}
	if _, dropErr := conn.Exec(ctx, "DROP TABLE public.period_approvals"); dropErr != nil {
		t.Fatalf("drop extra table: %v", dropErr)
	}

	// 3. Extra column detection
	if _, alterErr := conn.Exec(ctx, "ALTER TABLE public.users ADD COLUMN unexpected_col TEXT"); alterErr != nil {
		t.Fatalf("add extra column: %v", alterErr)
	}
	valErr = validateBaselineSchema(ctx, conn)
	if valErr == nil || !strings.Contains(valErr.Error(), "unexpected extra column users.unexpected_col") {
		t.Fatalf("expected unexpected extra column error, got: %v", valErr)
	}
	if _, dropColErr := conn.Exec(ctx, "ALTER TABLE public.users DROP COLUMN unexpected_col"); dropColErr != nil {
		t.Fatalf("drop extra column: %v", dropColErr)
	}

	// 4. Missing column / type divergence
	if _, renameErr := conn.Exec(ctx, "ALTER TABLE public.users RENAME COLUMN name TO legacy_name"); renameErr != nil {
		t.Fatalf("rename column: %v", renameErr)
	}
	valErr = validateBaselineSchema(ctx, conn)
	if valErr == nil || !strings.Contains(valErr.Error(), "missing column users.name") {
		t.Fatalf("expected missing column error, got: %v", valErr)
	}
	if _, restoreErr := conn.Exec(ctx, "ALTER TABLE public.users RENAME COLUMN legacy_name TO name"); restoreErr != nil {
		t.Fatalf("restore column: %v", restoreErr)
	}

	// 5. Column nullability divergence
	if _, dropNotnullErr := conn.Exec(ctx, "ALTER TABLE public.users ALTER COLUMN name DROP NOT NULL"); dropNotnullErr != nil {
		t.Fatalf("drop not null: %v", dropNotnullErr)
	}
	valErr = validateBaselineSchema(ctx, conn)
	if valErr == nil || !strings.Contains(valErr.Error(), "unexpected nullable status") {
		t.Fatalf("expected nullability divergence error, got: %v", valErr)
	}
	if _, setNotnullErr := conn.Exec(ctx, "ALTER TABLE public.users ALTER COLUMN name SET NOT NULL"); setNotnullErr != nil {
		t.Fatalf("restore not null: %v", setNotnullErr)
	}

	// 6. Schema is valid again
	if cleanErr := validateBaselineSchema(ctx, conn); cleanErr != nil {
		t.Fatalf("expected schema to pass validation after restoration: %v", cleanErr)
	}
}

func TestReconcileBaseline_Integration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("skipping reconcile baseline integration test: DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	conn, acquireErr := pool.Acquire(ctx)
	if acquireErr != nil {
		t.Fatalf("acquire conn: %v", acquireErr)
	}
	defer conn.Release()

	// Ensure migrations table exists
	if ensureErr := ensureMigrationsTable(ctx, conn); ensureErr != nil {
		t.Fatalf("ensure table: %v", ensureErr)
	}

	// Create temp dir with baseline file
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, baselineFileName)
	if writeErr := os.WriteFile(baselinePath, []byte("-- dummy baseline"), 0644); writeErr != nil {
		t.Fatalf("write baseline: %v", writeErr)
	}

	// Clear baseline row if present
	_, _ = conn.Exec(ctx, "DELETE FROM schema_migrations WHERE filename = $1", baselineFileName)

	// Simulate 33 legacy migrations
	for i := 1; i <= 32; i++ {
		_, _ = conn.Exec(ctx, "INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			filepath.Join(tmpDir, filepath.Base(os.Args[0]))+string(rune('0'+i)), "dummy")
	}
	_, _ = conn.Exec(ctx, "INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		finalLegacyMigrationFile, "dummy")

	defer func() {
		_, _ = conn.Exec(ctx, "DELETE FROM schema_migrations")
	}()

	// 1. With divergent schema (extra table): reconcileBaseline MUST fail and NOT insert baseline
	if _, createErr := conn.Exec(ctx, "CREATE TABLE public.period_approvals (id TEXT PRIMARY KEY)"); createErr != nil {
		t.Fatalf("create extra table: %v", createErr)
	}
	recErr := reconcileBaseline(ctx, conn, tmpDir)
	if recErr == nil || !strings.Contains(recErr.Error(), "validate baseline schema") {
		t.Fatalf("expected validate baseline schema error, got: %v", recErr)
	}

	var count int
	if scanErr := conn.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", baselineFileName).Scan(&count); scanErr != nil {
		t.Fatalf("query count: %v", scanErr)
	}
	if count != 0 {
		t.Fatalf("expected 0 baseline records on divergent schema, got %d", count)
	}

	// Drop extra table to restore valid schema
	if _, dropErr := conn.Exec(ctx, "DROP TABLE public.period_approvals"); dropErr != nil {
		t.Fatalf("drop extra table: %v", dropErr)
	}

	// 2. With valid schema: reconcileBaseline succeeds and inserts baseline record
	if recOkErr := reconcileBaseline(ctx, conn, tmpDir); recOkErr != nil {
		t.Fatalf("expected reconcileBaseline to succeed on valid schema: %v", recOkErr)
	}

	if scanErr := conn.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", baselineFileName).Scan(&count); scanErr != nil {
		t.Fatalf("query count: %v", scanErr)
	}
	if count != 1 {
		t.Fatalf("expected 1 baseline record after successful reconciliation, got %d", count)
	}
}
