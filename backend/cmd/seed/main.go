package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("SEED_ENVIRONMENT")
	}
	if env == "production" || env == "prod" {
		return errors.New("security error: database seeding is strictly prohibited in production")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := seedDemoData(ctx, tx); err != nil {
		return fmt.Errorf("seed demo data: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	slog.Info("demo seeds successfully applied (login: demo@micha.local / password123)")
	return nil
}

func seedDemoData(ctx context.Context, tx pgx.Tx) error {
	now := time.Now().UTC()

	// 1. Demo User
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	userID := uuid.NewString()
	var actualUserID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, name, created_at)
		VALUES ($1, 'demo@micha.local', $2, 'Carlos Demo', $3)
		ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, userID, string(passwordHash), now).Scan(&actualUserID)
	if err != nil {
		return fmt.Errorf("insert demo user: %w", err)
	}

	// 2. Demo Household
	householdID := uuid.NewString()
	var actualHouseholdID string
	err = tx.QueryRow(ctx, `
		INSERT INTO households (id, name, owner_id, settlement_mode, currency, closing_day, period_frequency, created_at, updated_at)
		VALUES ($1, 'Hogar Demo', $2, 'equal', 'MXN', 15, 'monthly', $3, $3)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, householdID, actualUserID, now).Scan(&actualHouseholdID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `SELECT id FROM households WHERE owner_id = $1 LIMIT 1`, actualUserID).Scan(&actualHouseholdID)
	}
	if err != nil {
		return fmt.Errorf("insert demo household: %w", err)
	}

	// 3. Demo Members (Carlos & Sofia)
	carlosID := uuid.NewString()
	var actualCarlosID string
	err = tx.QueryRow(ctx, `
		INSERT INTO members (id, household_id, user_id, name, email, monthly_salary_cents, created_at, updated_at)
		VALUES ($1, $2, $3, 'Carlos Demo', 'demo@micha.local', 3000000, $4, $4)
		ON CONFLICT (household_id, email) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, carlosID, actualHouseholdID, actualUserID, now).Scan(&actualCarlosID)
	if err != nil {
		return fmt.Errorf("insert member Carlos: %w", err)
	}

	sofiaID := uuid.NewString()
	var actualSofiaID string
	err = tx.QueryRow(ctx, `
		INSERT INTO members (id, household_id, name, email, monthly_salary_cents, created_at, updated_at)
		VALUES ($1, $2, 'Sofia Demo', 'sofia@micha.local', 2500000, $3, $3)
		ON CONFLICT (household_id, email) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, sofiaID, actualHouseholdID, now).Scan(&actualSofiaID)
	if err != nil {
		return fmt.Errorf("insert member Sofia: %w", err)
	}

	// 4. Split Config (50/50)
	for _, memberID := range []string{actualCarlosID, actualSofiaID} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO household_split_config (household_id, member_id, percentage)
			VALUES ($1, $2, 50.00)
			ON CONFLICT (household_id, member_id) DO UPDATE SET percentage = 50.00
		`, actualHouseholdID, memberID); err != nil {
			return fmt.Errorf("insert split config: %w", err)
		}
	}

	// 5. Default Categories
	categories := []struct {
		name string
		slug string
	}{
		{"Rent", "rent"},
		{"Auto", "auto"},
		{"Streaming", "streaming"},
		{"Food", "food"},
		{"Personal", "personal"},
		{"Savings", "savings"},
		{"Other", "other"},
	}

	categoryIDs := make(map[string]string)
	for _, c := range categories {
		catID := uuid.NewString()
		var actualCatID string
		err := tx.QueryRow(ctx, `
			INSERT INTO categories (id, household_id, name, slug, is_default, created_at)
			VALUES ($1, $2, $3, $4, true, $5)
			ON CONFLICT (household_id, slug) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, catID, actualHouseholdID, c.name, c.slug, now).Scan(&actualCatID)
		if err != nil {
			return fmt.Errorf("insert category %s: %w", c.slug, err)
		}
		categoryIDs[c.slug] = actualCatID
	}

	// 6. Current Period
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	periodID := uuid.NewString()
	var actualPeriodID string
	err = tx.QueryRow(ctx, `
		INSERT INTO periods (id, household_id, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'open', $5, $5)
		ON CONFLICT (household_id, id) DO NOTHING
		RETURNING id
	`, periodID, actualHouseholdID, startDate, endDate, now).Scan(&actualPeriodID)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.QueryRow(ctx, `SELECT id FROM periods WHERE household_id = $1 LIMIT 1`, actualHouseholdID).Scan(&actualPeriodID)
	} else if err != nil {
		return fmt.Errorf("insert demo period: %w", err)
	}

	// 7. Cards
	cardID := uuid.NewString()
	var actualCardID string
	err = tx.QueryRow(ctx, `
		INSERT INTO cards (id, household_id, owner_member_id, bank_name, card_name, cutoff_day, created_at, updated_at)
		VALUES ($1, $2, $3, 'BBVA', 'BBVA Azul Débito', 15, $4, $4)
		ON CONFLICT (household_id, id) DO NOTHING
		RETURNING id
	`, cardID, actualHouseholdID, actualCarlosID, now).Scan(&actualCardID)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.QueryRow(ctx, `SELECT id FROM cards WHERE household_id = $1 LIMIT 1`, actualHouseholdID).Scan(&actualCardID)
	} else if err != nil {
		return fmt.Errorf("insert demo card: %w", err)
	}

	// 8. Sample Expenses
	sampleExpenses := []struct {
		desc        string
		amountCents int64
		catSlug     string
		memberID    string
		method      string
	}{
		{"Supermercado semanal", 145000, "food", actualCarlosID, "card"},
		{"Renta mensual departamento", 1200000, "rent", actualCarlosID, "transfer"},
		{"Gasolina tanque lleno", 85000, "auto", actualSofiaID, "card"},
		{"Cena fin de semana", 62000, "food", actualSofiaID, "cash"},
	}

	for _, se := range sampleExpenses {
		if _, err := tx.Exec(ctx, `
			INSERT INTO expenses (
				id, household_id, paid_by_member_id, period_id, category_id, card_id,
				amount_cents, description, is_shared, currency, payment_method, expense_type,
				created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, 'MXN', $9, 'variable', $10, $10)
		`, uuid.NewString(), actualHouseholdID, se.memberID, actualPeriodID, categoryIDs[se.catSlug], actualCardID, se.amountCents, se.desc, se.method, now); err != nil {
			return fmt.Errorf("insert expense %q: %w", se.desc, err)
		}
	}

	// 9. Sample Recurring Expense (Streaming)
	recurringID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO recurring_expenses (
			id, household_id, paid_by_member_id, amount_cents, description, category_id,
			expense_type, recurrence_pattern, start_date, next_generation_date, is_active, is_agnostic,
			created_at, updated_at
		)
		VALUES ($1, $2, NULL, 26900, 'Netflix Estándar', $3, 'fixed', 'monthly', $4, $4, true, true, $5, $5)
		ON CONFLICT (id) DO NOTHING
	`, recurringID, actualHouseholdID, categoryIDs["streaming"], startDate, now); err != nil {
		return fmt.Errorf("insert recurring expense: %w", err)
	}

	// Link with catalog service if exists
	var catalogServiceID string
	if err := tx.QueryRow(ctx, `SELECT id FROM subscription_services WHERE slug = 'netflix-standard' LIMIT 1`).Scan(&catalogServiceID); err == nil && catalogServiceID != "" {
		_, _ = tx.Exec(ctx, `
			INSERT INTO recurring_expense_catalog_links (recurring_expense_id, catalog_service_id, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT DO NOTHING
		`, recurringID, catalogServiceID, now)
	}

	return nil
}
