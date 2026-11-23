package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// createIsolatedTestDB creates an isolated temporary PostgreSQL database
// for running destructive integration tests without risking the shared development or CI database.
// It returns a pgxpool.Pool connected to the isolated database and a cleanup function.
func createIsolatedTestDB(t *testing.T, ctx context.Context, baseDatabaseURL string) (*pgxpool.Pool, func()) {
	t.Helper()

	cfg, parseErr := pgxpool.ParseConfig(baseDatabaseURL)
	if parseErr != nil {
		t.Fatalf("parse database url: %v", parseErr)
	}

	adminPool, adminErr := pgxpool.New(ctx, baseDatabaseURL)
	if adminErr != nil {
		t.Fatalf("connect to base database: %v", adminErr)
	}

	testDBName := fmt.Sprintf("test_micha_%d", time.Now().UnixNano()%100000000)
	if _, createErr := adminPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName)); createErr != nil {
		adminPool.Close()
		t.Skipf("skipping test: database user lacks CREATE DATABASE privilege: %v", createErr)
		return nil, func() {}
	}

	cfg.ConnConfig.Database = testDBName
	testPool, poolErr := pgxpool.NewWithConfig(ctx, cfg)
	if poolErr != nil {
		_, _ = adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", testDBName))
		adminPool.Close()
		t.Fatalf("connect to test database: %v", poolErr)
	}

	cleanup := func() {
		testPool.Close()
		if _, dropErr := adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", testDBName)); dropErr != nil {
			t.Logf("cleanup: drop test database %s: %v", testDBName, dropErr)
		}
		adminPool.Close()
	}

	return testPool, cleanup
}

func TestApply_Integration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("skipping migration integration test: DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, cleanup := createIsolatedTestDB(t, ctx, databaseURL)
	defer cleanup()

	tmpDir := t.TempDir()
	mig1 := filepath.Join(tmpDir, "001_test_migration.sql")
	if writeErr := os.WriteFile(mig1, []byte("CREATE TABLE IF NOT EXISTS _test_table (id TEXT PRIMARY KEY);"), 0644); writeErr != nil {
		t.Fatalf("failed to write mig1: %v", writeErr)
	}

	if applyErr := Apply(ctx, pool, tmpDir); applyErr != nil {
		t.Fatalf("Apply failed: %v", applyErr)
	}

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

// setupLegacySchema creates the full database schema matching the pre-consolidation
// legacy state (33 migrations ending in 033_add_name_to_users.sql) inside an isolated test database.
// This establishes: household_split_config.percentage as integer with legacy CHECK (percentage > 0),
// installments columns (id, expense_id, paid_by_member_id as varchar(255), created_at/updated_at as timestamp,
// card_id text), and installments.paid_by_member_id foreign key without cascade (NO ACTION).
func setupLegacySchema(t *testing.T, ctx context.Context, conn *pgxpool.Conn) {
	t.Helper()

	schemaSQL, readErr := os.ReadFile(filepath.Join("..", "..", "..", "migrations", baselineFileName))
	if readErr != nil {
		t.Fatalf("read baseline schema file: %v", readErr)
	}
	if _, execErr := conn.Exec(ctx, string(schemaSQL)); execErr != nil {
		t.Fatalf("init base schema: %v", execErr)
	}

	_, alterErr := conn.Exec(ctx, `
		ALTER TABLE public.household_split_config DROP CONSTRAINT IF EXISTS household_split_config_percentage_check;
		ALTER TABLE public.household_split_config ALTER COLUMN percentage TYPE integer;
		ALTER TABLE public.household_split_config ADD CONSTRAINT household_split_config_percentage_check CHECK (percentage > 0 AND percentage <= 100);

		ALTER TABLE public.installments ADD COLUMN IF NOT EXISTS card_id text;
		ALTER TABLE public.installments ALTER COLUMN id TYPE varchar(255);
		ALTER TABLE public.installments ALTER COLUMN expense_id TYPE varchar(255);
		ALTER TABLE public.installments DROP CONSTRAINT IF EXISTS installments_paid_by_member_id_fkey;
		ALTER TABLE public.installments ALTER COLUMN paid_by_member_id TYPE varchar(255);
		ALTER TABLE public.installments ADD CONSTRAINT installments_paid_by_member_id_fkey FOREIGN KEY (paid_by_member_id) REFERENCES public.members(id);
		ALTER TABLE public.installments ALTER COLUMN created_at TYPE timestamp USING created_at AT TIME ZONE 'UTC';
		ALTER TABLE public.installments ALTER COLUMN updated_at TYPE timestamp USING updated_at AT TIME ZONE 'UTC';
	`)
	if alterErr != nil {
		t.Fatalf("setup legacy schema: %v", alterErr)
	}
}

func TestValidateBaselineSchema_Integration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("skipping schema validation integration test: DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, cleanup := createIsolatedTestDB(t, ctx, databaseURL)
	defer cleanup()

	conn, acquireErr := pool.Acquire(ctx)
	if acquireErr != nil {
		t.Fatalf("acquire conn: %v", acquireErr)
	}
	defer conn.Release()

	// 1. Valid canonical baseline schema check (from 001_initial_schema.sql)
	schemaSQL, readErr := os.ReadFile(filepath.Join("..", "..", "..", "migrations", baselineFileName))
	if readErr != nil {
		t.Fatalf("read baseline schema file: %v", readErr)
	}
	if _, execErr := conn.Exec(ctx, string(schemaSQL)); execErr != nil {
		t.Fatalf("init base schema: %v", execErr)
	}
	if valErr := validateBaselineSchema(ctx, conn); valErr != nil {
		t.Fatalf("expected valid canonical baseline schema to pass validation: %v", valErr)
	}

	// 2. Valid legacy baseline schema check (int4 percentage, varchar/timestamp/card_id in installments, legacy constraints)
	if _, alterErr := conn.Exec(ctx, `
		ALTER TABLE public.household_split_config DROP CONSTRAINT IF EXISTS household_split_config_percentage_check;
		ALTER TABLE public.household_split_config ALTER COLUMN percentage TYPE integer;
		ALTER TABLE public.household_split_config ADD CONSTRAINT household_split_config_percentage_check CHECK (percentage > 0 AND percentage <= 100);
		ALTER TABLE public.installments ADD COLUMN IF NOT EXISTS card_id text;
		ALTER TABLE public.installments ALTER COLUMN id TYPE varchar(255);
		ALTER TABLE public.installments ALTER COLUMN expense_id TYPE varchar(255);
		ALTER TABLE public.installments DROP CONSTRAINT IF EXISTS installments_paid_by_member_id_fkey;
		ALTER TABLE public.installments ALTER COLUMN paid_by_member_id TYPE varchar(255);
		ALTER TABLE public.installments ADD CONSTRAINT installments_paid_by_member_id_fkey FOREIGN KEY (paid_by_member_id) REFERENCES public.members(id);
		ALTER TABLE public.installments ALTER COLUMN created_at TYPE timestamp USING created_at AT TIME ZONE 'UTC';
		ALTER TABLE public.installments ALTER COLUMN updated_at TYPE timestamp USING updated_at AT TIME ZONE 'UTC';
	`); alterErr != nil {
		t.Fatalf("alter to legacy schema: %v", alterErr)
	}
	if valErr := validateBaselineSchema(ctx, conn); valErr != nil {
		t.Fatalf("expected valid legacy baseline schema to pass validation: %v", valErr)
	}

	// 3. Extra table detection (e.g. obsolete table period_approvals)
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

	// 4. Extra column detection
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

	// 5. Missing column
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

	// 6. Column nullability divergence
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

	// 7. Invalid column type (e.g. users.name changed to boolean, which is neither text nor varchar)
	if _, typeErr := conn.Exec(ctx, "ALTER TABLE public.users ALTER COLUMN name TYPE boolean USING false"); typeErr != nil {
		t.Fatalf("alter name to boolean: %v", typeErr)
	}
	valErr = validateBaselineSchema(ctx, conn)
	if valErr == nil || !strings.Contains(valErr.Error(), "has unexpected type") {
		t.Fatalf("expected type divergence error for name, got: %v", valErr)
	}
	if _, restoreTypeErr := conn.Exec(ctx, "ALTER TABLE public.users ALTER COLUMN name TYPE text USING ''"); restoreTypeErr != nil {
		t.Fatalf("restore name to text: %v", restoreTypeErr)
	}

	// 8. Schema is valid again
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
	pool, cleanup := createIsolatedTestDB(t, ctx, databaseURL)
	defer cleanup()

	conn, acquireErr := pool.Acquire(ctx)
	if acquireErr != nil {
		t.Fatalf("acquire conn: %v", acquireErr)
	}
	defer conn.Release()

	// Ensure migrations table exists
	if ensureErr := ensureMigrationsTable(ctx, conn); ensureErr != nil {
		t.Fatalf("ensure table: %v", ensureErr)
	}

	// Reconstruct the authentic legacy 033 schema (varchar/timestamp/percentage int4/installments.card_id/legacy constraints)
	setupLegacySchema(t, ctx, conn)

	// Populate test data to verify legacy constraint behavior prior to reconciliation
	if _, seedErr := conn.Exec(ctx, `
		INSERT INTO public.households (id, name, settlement_mode, currency, closing_day, period_frequency)
		VALUES ('hh-reconcile-1', 'Legacy Household', 'equal', 'MXN', 1, 'monthly');
		INSERT INTO public.members (id, household_id, name, email, monthly_salary_cents)
		VALUES ('m-reconcile-1', 'hh-reconcile-1', 'Legacy Member', 'legacy@example.com', 500000);
	`); seedErr != nil {
		t.Fatalf("insert test household and member: %v", seedErr)
	}

	// Legacy CHECK verification: inserting percentage = 0 must fail on authentic legacy schema
	if _, insert0Err := conn.Exec(ctx, `
		INSERT INTO public.household_split_config (household_id, member_id, percentage)
		VALUES ('hh-reconcile-1', 'm-reconcile-1', 0)
	`); insert0Err == nil {
		t.Fatalf("expected legacy CHECK constraint (percentage > 0) to reject percentage = 0, but insert succeeded")
	}

	// Create temp dir with baseline file
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, baselineFileName)
	if writeErr := os.WriteFile(baselinePath, []byte("-- dummy baseline"), 0644); writeErr != nil {
		t.Fatalf("write baseline: %v", writeErr)
	}

	// Simulate 33 legacy migrations with explicit test filenames
	testLegacyFiles := make([]string, 0, 33)
	for i := 1; i <= 32; i++ {
		fname := fmt.Sprintf("test_legacy_migration_%02d.sql", i)
		testLegacyFiles = append(testLegacyFiles, fname)
		if _, insertErr := conn.Exec(ctx, "INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2)", fname, "dummy"); insertErr != nil {
			t.Fatalf("insert test legacy migration %s: %v", fname, insertErr)
		}
	}
	testLegacyFiles = append(testLegacyFiles, finalLegacyMigrationFile)
	if _, insertErr := conn.Exec(ctx, "INSERT INTO schema_migrations (filename, checksum) VALUES ($1, $2)", finalLegacyMigrationFile, "dummy"); insertErr != nil {
		t.Fatalf("insert final legacy migration: %v", insertErr)
	}

	// Scoped defer cleanup: clean only test-specific legacy rows
	defer func() {
		for _, f := range testLegacyFiles {
			_, _ = conn.Exec(ctx, "DELETE FROM schema_migrations WHERE filename = $1", f)
		}
		_, _ = conn.Exec(ctx, "DELETE FROM schema_migrations WHERE filename = $1", baselineFileName)
	}()

	// 1. With divergent schema (extra table period_approvals): reconcileBaseline MUST fail and NOT insert baseline
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

	// Drop extra table to restore valid legacy schema
	if _, dropErr := conn.Exec(ctx, "DROP TABLE public.period_approvals"); dropErr != nil {
		t.Fatalf("drop extra table: %v", dropErr)
	}

	// 2. With valid authentic legacy schema: reconcileBaseline succeeds, inserts baseline record, and aligns schema
	if recOkErr := reconcileBaseline(ctx, conn, tmpDir); recOkErr != nil {
		t.Fatalf("expected reconcileBaseline to succeed on valid legacy schema: %v", recOkErr)
	}

	if scanErr := conn.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", baselineFileName).Scan(&count); scanErr != nil {
		t.Fatalf("query count: %v", scanErr)
	}
	if count != 1 {
		t.Fatalf("expected 1 baseline record after successful reconciliation, got %d", count)
	}

	// 3. Verify that reconcileBaseline aligned legacy types to canonical v1 schema
	var pctType, idType, createdAtType string
	if scanErr := conn.QueryRow(ctx, `
		SELECT udt_name FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'household_split_config' AND column_name = 'percentage'
	`).Scan(&pctType); scanErr != nil {
		t.Fatalf("query percentage type: %v", scanErr)
	}
	if pctType != "numeric" {
		t.Errorf("expected percentage type to be aligned to 'numeric', got %q", pctType)
	}

	if scanErr := conn.QueryRow(ctx, `
		SELECT udt_name FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'installments' AND column_name = 'id'
	`).Scan(&idType); scanErr != nil {
		t.Fatalf("query installments.id type: %v", scanErr)
	}
	if idType != "text" {
		t.Errorf("expected installments.id type to be aligned to 'text', got %q", idType)
	}

	if scanErr := conn.QueryRow(ctx, `
		SELECT udt_name FROM information_schema.columns 
		WHERE table_schema = 'public' AND table_name = 'installments' AND column_name = 'created_at'
	`).Scan(&createdAtType); scanErr != nil {
		t.Fatalf("query installments.created_at type: %v", scanErr)
	}
	if createdAtType != "timestamptz" {
		t.Errorf("expected installments.created_at type to be aligned to 'timestamptz', got %q", createdAtType)
	}

	// Verify installments.card_id is still present and was not dropped
	var cardIdExists bool
	if scanErr := conn.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns 
			WHERE table_schema = 'public' AND table_name = 'installments' AND column_name = 'card_id'
		)
	`).Scan(&cardIdExists); scanErr != nil {
		t.Fatalf("check card_id exists: %v", scanErr)
	}
	if !cardIdExists {
		t.Errorf("expected installments.card_id to be preserved after reconciliation")
	}

	// 4. Verify that constraints have been aligned to canonical v1:
	// 4a. Canonical CHECK constraint on household_split_config now allows percentage = 0
	if _, insert0AfterErr := conn.Exec(ctx, `
		INSERT INTO public.household_split_config (household_id, member_id, percentage)
		VALUES ('hh-reconcile-1', 'm-reconcile-1', 0)
	`); insert0AfterErr != nil {
		t.Fatalf("expected canonical CHECK to allow percentage = 0 after reconciliation: %v", insert0AfterErr)
	}

	// 4b. Canonical FK cascade on installments: deleting member deletes installments
	if _, seedExpErr := conn.Exec(ctx, `
		INSERT INTO public.categories (id, household_id, name, slug)
		VALUES ('cat-reconcile-1', 'hh-reconcile-1', 'Reconcile Cat', 'reconcile-cat');
		INSERT INTO public.expenses (id, household_id, paid_by_member_id, category_id, amount_cents, description)
		VALUES ('exp-reconcile-1', 'hh-reconcile-1', 'm-reconcile-1', 'cat-reconcile-1', 2000, 'Reconcile Exp');
		INSERT INTO public.installments (id, expense_id, paid_by_member_id, start_date, installment_amount_cents, total_amount_cents, total_installments, current_installment)
		VALUES ('inst-reconcile-1', 'exp-reconcile-1', 'm-reconcile-1', CURRENT_DATE, 1000, 2000, 2, 1);
	`); seedExpErr != nil {
		t.Fatalf("insert test expense and installment: %v", seedExpErr)
	}

	if _, delMemberErr := conn.Exec(ctx, `DELETE FROM public.members WHERE id = 'm-reconcile-1'`); delMemberErr != nil {
		t.Fatalf("expected deleting member to cascade to installments without FK violation: %v", delMemberErr)
	}

	var remainingInstallments int
	if scanErr := conn.QueryRow(ctx, `SELECT COUNT(*) FROM public.installments WHERE id = 'inst-reconcile-1'`).Scan(&remainingInstallments); scanErr != nil {
		t.Fatalf("query remaining installments count: %v", scanErr)
	}
	if remainingInstallments != 0 {
		t.Fatalf("expected installment to be deleted by cascade, but found %d", remainingInstallments)
	}

	// 5. Verify that validateBaselineSchema passes on the reconciled database
	if valErr := validateBaselineSchema(ctx, conn); valErr != nil {
		t.Fatalf("expected validateBaselineSchema to pass on reconciled schema: %v", valErr)
	}
}
