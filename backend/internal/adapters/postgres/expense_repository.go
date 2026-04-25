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

// ListByHouseholdAndPeriod returns non-deleted household expenses between [from, to).
func (r ExpenseRepository) ListByHouseholdAndPeriod(ctx context.Context, householdID string, from, to time.Time) ([]expense.Expense, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, paid_by_member_id, amount_cents, description, is_shared, currency, payment_method, expense_type, card_id, card_name, category_id, total_installments, created_at, updated_at, deleted_at
			FROM expenses
			WHERE household_id = $1
				AND created_at >= $2
				AND created_at < $3
				AND deleted_at IS NULL
			ORDER BY created_at ASC`,
		householdID, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("expense repository listByHouseholdAndPeriod: %w", err)
	}
	defer rows.Close()

	expenses := make([]expense.Expense, 0)
	for rows.Next() {
		e, scanErr := scanExpense(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("expense repository listByHouseholdAndPeriod: scan: %w", scanErr)
		}
		expenses = append(expenses, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("expense repository listByHouseholdAndPeriod: rows: %w", err)
	}

	return expenses, nil
}

// NewExpenseRepository constructs an ExpenseRepository backed by the given pool.
func NewExpenseRepository(db *pgxpool.Pool) ExpenseRepository {
	return ExpenseRepository{db: db}
}

// Save persists a new expense record.
func (r ExpenseRepository) Save(ctx context.Context, e expense.Expense) error {
	attrs := e.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO expenses (id, household_id, paid_by_member_id, amount_cents, description, is_shared, currency, payment_method, expense_type, card_id, card_name, category_id, total_installments, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		string(attrs.ID), attrs.HouseholdID, nullIfEmpty(attrs.PaidByMemberID), attrs.AmountCents,
		attrs.Description, attrs.IsShared, attrs.Currency, string(attrs.PaymentMethod),
		string(attrs.ExpenseType), nullIfEmpty(attrs.CardID), attrs.CardName, attrs.CategoryID,
		attrs.TotalInstallments, attrs.CreatedAt, attrs.UpdatedAt,
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
		`SELECT id, household_id, paid_by_member_id, amount_cents, description, is_shared, currency, payment_method, expense_type, card_id, card_name, category_id, total_installments, created_at, updated_at, deleted_at
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
		`SELECT id, household_id, paid_by_member_id, amount_cents, description, is_shared, currency, payment_method, expense_type, card_id, card_name, category_id, total_installments, created_at, updated_at, deleted_at
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
		SET paid_by_member_id = $1,
			amount_cents      = $2,
			description       = $3,
			is_shared         = $4,
			currency          = $5,
			payment_method    = $6,
			expense_type      = $7,
			card_id           = $8,
			card_name         = $9,
			category_id       = $10,
			total_installments = $11,
			updated_at        = $12,
			deleted_at        = $13
		WHERE id = $14`,
		nullIfEmpty(attrs.PaidByMemberID), attrs.AmountCents, attrs.Description,
		attrs.IsShared, attrs.Currency, string(attrs.PaymentMethod),
		string(attrs.ExpenseType), nullIfEmpty(attrs.CardID), attrs.CardName, attrs.CategoryID,
		attrs.TotalInstallments, attrs.UpdatedAt, attrs.DeletedAt, string(attrs.ID),
	)
	if err != nil {
		return fmt.Errorf("expense repository update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// SumPersonalByMemberAndPeriod returns personal outflow for one member in [from, to).
func (r ExpenseRepository) SumPersonalByMemberAndPeriod(ctx context.Context, householdID, memberID string, from, to time.Time) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx,
		`SELECT
			COALESCE((
				SELECT SUM(e.amount_cents)
				FROM expenses e
				JOIN categories c ON c.id = e.category_id
				WHERE e.household_id = $1
					AND e.paid_by_member_id = $2
					AND e.deleted_at IS NULL
					AND e.is_shared = FALSE
					AND c.slug = 'personal'
					AND e.expense_type <> 'msi'
					AND e.created_at >= $3
					AND e.created_at < $4
			), 0)
			+
			COALESCE((
				SELECT SUM(i.installment_amount_cents)
				FROM installments i
				JOIN expenses e ON e.id = i.expense_id
				JOIN categories c ON c.id = e.category_id
				WHERE e.household_id = $1
					AND i.paid_by_member_id = $2
					AND e.deleted_at IS NULL
					AND e.is_shared = FALSE
					AND e.expense_type = 'msi'
					AND c.slug = 'personal'
					AND i.start_date >= $3
					AND i.start_date < $4
			), 0)`,
		householdID, memberID, from, to,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("expense repository sumPersonalByMemberAndPeriod: %w", err)
	}

	return total, nil
}

// ensure interface compliance at compile time.
var _ outbound.ExpenseRepository = ExpenseRepository{}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// row is the minimal interface shared by pgx.Row and pgx.Rows.
type row interface {
	Scan(dest ...any) error
}

func scanExpense(r row) (expense.Expense, error) {
	var (
		id                string
		householdID       string
		paidByMemberID    *string
		amountCents       int64
		description       string
		isShared          bool
		currency          string
		paymentMethod     string
		expenseType       string
		cardID            *string
		cardName          string
		categoryID        string
		totalInstallments int
		createdAt         time.Time
		updatedAt         time.Time
		deletedAt         *time.Time
	)

	if err := r.Scan(&id, &householdID, &paidByMemberID, &amountCents, &description, &isShared, &currency, &paymentMethod, &expenseType, &cardID, &cardName, &categoryID, &totalInstallments, &createdAt, &updatedAt, &deletedAt); err != nil {
		return expense.Expense{}, err
	}

	cID := ""
	if cardID != nil {
		cID = *cardID
	}

	return expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:                expense.ID(id),
		HouseholdID:       householdID,
		PaidByMemberID:    valueOrEmptyString(paidByMemberID),
		AmountCents:       amountCents,
		Description:       description,
		IsShared:          isShared,
		Currency:          currency,
		PaymentMethod:     expense.PaymentMethod(paymentMethod),
		ExpenseType:       expense.ExpenseType(expenseType),
		CardID:            cID,
		CardName:          cardName,
		CategoryID:        categoryID,
		TotalInstallments: totalInstallments,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		DeletedAt:         deletedAt,
	})
}

func valueOrEmptyString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
