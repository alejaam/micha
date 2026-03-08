// Package postgres provides PostgreSQL adapter implementations for outbound ports.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

// ExpenseRepository fulfils outbound.ExpenseRepository using PostgreSQL.
type ExpenseRepository struct {
	db *pgxpool.Pool
}

// NewExpenseRepository constructs an ExpenseRepository backed by the given pool.
func NewExpenseRepository(db *pgxpool.Pool) ExpenseRepository {
	return ExpenseRepository{db: db}
}

// Save persists a new expense record.
func (r ExpenseRepository) Save(ctx context.Context, e expense.Expense) error {
	attrs := e.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO expenses (id, household_id, amount_cents, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
		string(attrs.ID), attrs.HouseholdID, attrs.AmountCents,
		attrs.Description, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("expense repository save: %w", err)
	}

	return nil
}

// FindByID retrieves an expense by ID. Returns shared.ErrNotFound when not found.
// Note: soft-deleted rows are still returned so callers can inspect DeletedAt.
func (r ExpenseRepository) FindByID(ctx context.Context, id string) (expense.Expense, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, amount_cents, description, created_at, updated_at, deleted_at
			FROM expenses
			WHERE id = $1`,
		id,
	)

	e, err := scanExpense(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return expense.Expense{}, shared.ErrNotFound
	}
	if err != nil {
		return expense.Expense{}, fmt.Errorf("expense repository findByID: %w", err)
	}

	return e, nil
}

// List returns non-deleted expenses for a household ordered by created_at DESC.
func (r ExpenseRepository) List(ctx context.Context, householdID string, limit, offset int) ([]expense.Expense, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, amount_cents, description, created_at, updated_at, deleted_at
			FROM expenses
			WHERE household_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`,
		householdID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("expense repository list: %w", err)
	}

	defer rows.Close()

	var expenses []expense.Expense
	for rows.Next() {
		e, err := scanExpense(rows)
		if err != nil {
			return nil, fmt.Errorf("expense repository list: scan: %w", err)
		}
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("expense repository list: rows: %w", err)
	}

	return expenses, nil
}

// Update persists changes to an existing, non-deleted expense.
func (r ExpenseRepository) Update(ctx context.Context, e expense.Expense) error {
	attrs := e.Attributes()
	tag, err := r.db.Exec(ctx,
		`UPDATE expenses
			SET amount_cents = $1,
				description  = $2,
				updated_at   = $3,
				deleted_at   = $4
			WHERE id = $5`,
		attrs.AmountCents, attrs.Description,
		attrs.UpdatedAt, attrs.DeletedAt,
		string(attrs.ID),
	)
	if err != nil {
		return fmt.Errorf("expense repository update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// ensure interface compliance at compile time.
var _ outbound.ExpenseRepository = ExpenseRepository{}

// row is the minimal interface shared by pgx.Row and pgx.Rows.
type row interface {
	Scan(dest ...any) error
}

func scanExpense(r row) (expense.Expense, error) {
	var (
		id          string
		householdID string
		amountCents int64
		description string
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
	)

	if err := r.Scan(&id, &householdID, &amountCents, &description, &createdAt, &updatedAt, &deletedAt); err != nil {
		return expense.Expense{}, err
	}

	return expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:          expense.ID(id),
		HouseholdID: householdID,
		AmountCents: amountCents,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
	})
}
