package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/micha?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// 1. Create User
	userID := uuid.NewString()
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	_, err = pool.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, created_at)
		 VALUES ($1, $2, $3, $4) ON CONFLICT (email) DO NOTHING`,
		userID, email, string(hashedPassword), time.Now(),
	)
	if err != nil {
		log.Printf("Warning: user creation failed or user exists: %v", err)
	}

	// Get user ID if it already existed
	err = pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		log.Fatalf("Failed to get user ID: %v", err)
	}

	// 2. Create Household
	householdID := uuid.NewString()
	_, err = pool.Exec(ctx,
		`INSERT INTO households (id, name, settlement_mode, currency, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		householdID, "Test Household", "equal", "MXN", time.Now(), time.Now(),
	)
	if err != nil {
		log.Fatalf("Failed to create household: %v", err)
	}

	// 3. Create Member
	memberID := uuid.NewString()
	_, err = pool.Exec(ctx,
		`INSERT INTO members (id, household_id, user_id, name, email, monthly_salary_cents, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		memberID, householdID, userID, "Test User", email, 5000000, time.Now(), time.Now(),
	)
	if err != nil {
		log.Fatalf("Failed to create member: %v", err)
	}

	// 4. Seed Categories
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

	for _, cat := range categories {
		_, err = pool.Exec(ctx,
			`INSERT INTO categories (id, household_id, name, slug, is_default, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (household_id, slug) DO NOTHING`,
			uuid.NewString(), householdID, cat.name, cat.slug, true, time.Now(),
		)
		if err != nil {
			log.Printf("Warning: category %s creation failed: %v", cat.slug, err)
		}
	}

	// 5. Create some expenses
	expenses := []struct {
		desc   string
		amount int64
		cat    string
	}{
		{"Grocery store", 125050, "food"},
		{"Netflix", 19900, "streaming"},
		{"Rent April", 1500000, "rent"},
	}

	for _, exp := range expenses {
		var catID string
		err = pool.QueryRow(ctx, "SELECT id FROM categories WHERE household_id = $1 AND slug = $2", householdID, exp.cat).Scan(&catID)
		if err != nil {
			log.Printf("Warning: failed to find category %s: %v", exp.cat, err)
			continue
		}

		_, err = pool.Exec(ctx,
			`INSERT INTO expenses (id, household_id, paid_by_member_id, category_id, amount_cents, description, is_shared, currency, payment_method, expense_type, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			uuid.NewString(), householdID, memberID, catID, exp.amount, exp.desc, true, "MXN", "card", "variable", time.Now(), time.Now(),
		)
		if err != nil {
			log.Printf("Warning: expense %s creation failed: %v", exp.desc, err)
		}
	}

	fmt.Println("Seed completed successfully!")
	fmt.Printf("User: %s / password: %s\n", email, password)
	fmt.Printf("Household ID: %s\n", householdID)
}
