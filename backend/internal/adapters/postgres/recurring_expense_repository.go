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
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

// RecurringExpenseRepository fulfils outbound.RecurringExpenseRepository using PostgreSQL.
type RecurringExpenseRepository struct {
	db *pgxpool.Pool
}

// NewRecurringExpenseRepository constructs a RecurringExpenseRepository backed by the given pool.
func NewRecurringExpenseRepository(db *pgxpool.Pool) RecurringExpenseRepository {
	return RecurringExpenseRepository{db: db}
}

// Save persists a new recurring expense record.
func (r RecurringExpenseRepository) Save(ctx context.Context, re recurringexpense.RecurringExpense) error {
	attrs := re.Attributes()
	paidByMemberID := any(attrs.PaidByMemberID)
	if attrs.IsAgnostic || attrs.PaidByMemberID == "" {
		paidByMemberID = nil
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO recurring_expenses (id, household_id, paid_by_member_id, is_agnostic, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, end_date, next_generation_date, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		string(attrs.ID), attrs.HouseholdID, paidByMemberID, attrs.IsAgnostic, attrs.AmountCents,
		attrs.Description, attrs.CategoryID, string(attrs.ExpenseType), string(attrs.RecurrencePattern),
		attrs.StartDate, attrs.EndDate, attrs.NextGenerationDate, attrs.IsActive, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("recurring expense repository save: %w", err)
	}

	return nil
}

// FindByID retrieves a recurring expense by ID. Returns shared.ErrNotFound when not found.
// Note: soft-deleted rows are still returned so callers can inspect DeletedAt.
func (r RecurringExpenseRepository) FindByID(ctx context.Context, id string) (recurringexpense.RecurringExpense, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, paid_by_member_id, is_agnostic, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, end_date, next_generation_date, is_active, created_at, updated_at, deleted_at
			FROM recurring_expenses
			WHERE id = $1`,
		id,
	)

	re, err := scanRecurringExpense(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return recurringexpense.RecurringExpense{}, shared.ErrNotFound
	}
	if err != nil {
		return recurringexpense.RecurringExpense{}, fmt.Errorf("recurring expense repository findByID: %w", err)
	}

	return re, nil
}

// List returns non-deleted recurring expenses for a household ordered by created_at DESC.
func (r RecurringExpenseRepository) List(ctx context.Context, householdID string, limit, offset int) ([]recurringexpense.RecurringExpense, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, paid_by_member_id, is_agnostic, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, end_date, next_generation_date, is_active, created_at, updated_at, deleted_at
			FROM recurring_expenses
			WHERE household_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`,
		householdID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("recurring expense repository list: %w", err)
	}

	defer rows.Close()

	var recurringExpenses []recurringexpense.RecurringExpense
	for rows.Next() {
		re, err := scanRecurringExpense(rows)
		if err != nil {
			return nil, fmt.Errorf("recurring expense repository list: scan: %w", err)
		}
		recurringExpenses = append(recurringExpenses, re)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recurring expense repository list: rows: %w", err)
	}

	return recurringExpenses, nil
}

// ListDueForGeneration returns active, non-deleted recurring expenses where next_generation_date <= asOfDate.
func (r RecurringExpenseRepository) ListDueForGeneration(ctx context.Context, asOfDate time.Time) ([]recurringexpense.RecurringExpense, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, paid_by_member_id, is_agnostic, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, end_date, next_generation_date, is_active, created_at, updated_at, deleted_at
			FROM recurring_expenses
			WHERE deleted_at IS NULL
				AND is_active = true
				AND is_agnostic = false
				AND next_generation_date <= $1
				AND (end_date IS NULL OR end_date >= $1)
			ORDER BY next_generation_date ASC`,
		asOfDate,
	)
	if err != nil {
		return nil, fmt.Errorf("recurring expense repository listDueForGeneration: %w", err)
	}

	defer rows.Close()

	var recurringExpenses []recurringexpense.RecurringExpense
	for rows.Next() {
		re, err := scanRecurringExpense(rows)
		if err != nil {
			return nil, fmt.Errorf("recurring expense repository listDueForGeneration: scan: %w", err)
		}
		recurringExpenses = append(recurringExpenses, re)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("recurring expense repository listDueForGeneration: rows: %w", err)
	}

	return recurringExpenses, nil
}

// Update persists changes to an existing recurring expense.
func (r RecurringExpenseRepository) Update(ctx context.Context, re recurringexpense.RecurringExpense) error {
	attrs := re.Attributes()
	paidByMemberID := any(attrs.PaidByMemberID)
	if attrs.IsAgnostic || attrs.PaidByMemberID == "" {
		paidByMemberID = nil
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE recurring_expenses
			SET paid_by_member_id     = $1,
				is_agnostic           = $2,
				amount_cents          = $3,
				description           = $4,
				category_id           = $5,
				expense_type          = $6,
				recurrence_pattern    = $7,
				start_date            = $8,
				end_date              = $9,
				next_generation_date  = $10,
				is_active             = $11,
				updated_at            = $12,
				deleted_at            = $13
			WHERE id = $14`,
		paidByMemberID, attrs.IsAgnostic, attrs.AmountCents, attrs.Description,
		attrs.CategoryID, string(attrs.ExpenseType), string(attrs.RecurrencePattern),
		attrs.StartDate, attrs.EndDate, attrs.NextGenerationDate, attrs.IsActive,
		attrs.UpdatedAt, attrs.DeletedAt, string(attrs.ID),
	)
	if err != nil {
		return fmt.Errorf("recurring expense repository update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// ensure interface compliance at compile time.
var _ outbound.RecurringExpenseRepository = RecurringExpenseRepository{}

// row is the minimal interface shared by pgx.Row and pgx.Rows.
type recurringExpenseRow interface {
	Scan(dest ...any) error
}

func scanRecurringExpense(r recurringExpenseRow) (recurringexpense.RecurringExpense, error) {
	var (
		id                 string
		householdID        string
		paidByMemberID     *string
		isAgnostic         bool
		amountCents        int64
		description        string
		categoryID         string
		expenseType        string
		recurrencePattern  string
		startDate          time.Time
		endDate            *time.Time
		nextGenerationDate time.Time
		isActive           bool
		createdAt          time.Time
		updatedAt          time.Time
		deletedAt          *time.Time
	)

	if err := r.Scan(&id, &householdID, &paidByMemberID, &isAgnostic, &amountCents, &description, &categoryID, &expenseType, &recurrencePattern, &startDate, &endDate, &nextGenerationDate, &isActive, &createdAt, &updatedAt, &deletedAt); err != nil {
		return recurringexpense.RecurringExpense{}, err
	}

	return recurringexpense.NewFromAttributes(recurringexpense.RecurringExpenseAttributes{
		ID:                 recurringexpense.ID(id),
		HouseholdID:        householdID,
		PaidByMemberID:     valueOrEmpty(paidByMemberID),
		IsAgnostic:         isAgnostic,
		AmountCents:        amountCents,
		Description:        description,
		CategoryID:         categoryID,
		ExpenseType:        expense.ExpenseType(expenseType),
		RecurrencePattern:  recurringexpense.RecurrencePattern(recurrencePattern),
		StartDate:          startDate,
		EndDate:            endDate,
		NextGenerationDate: nextGenerationDate,
		IsActive:           isActive,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		DeletedAt:          deletedAt,
	})
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
