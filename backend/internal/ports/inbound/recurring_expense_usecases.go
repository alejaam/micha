package inbound

import (
	"context"
	"time"

	"micha/backend/internal/domain/recurringexpense"
)

// CreateRecurringExpenseInput contains required data to create a recurring expense.
type CreateRecurringExpenseInput struct {
	HouseholdID       string
	PaidByMemberID    string
	AmountCents       int64
	Description       string
	CategoryID        string
	ExpenseType       string
	RecurrencePattern string
	StartDate         time.Time
	EndDate           *time.Time
}

// CreateRecurringExpenseOutput contains created recurring expense identifiers.
type CreateRecurringExpenseOutput struct {
	RecurringExpenseID string
}

// ListRecurringExpensesQuery holds the parameters for listing recurring expenses.
type ListRecurringExpensesQuery struct {
	HouseholdID string
	Limit       int
	Offset      int
}

// UpdateRecurringExpenseCommand holds the partial-update fields for a recurring expense.
// Only non-nil fields are applied.
type UpdateRecurringExpenseCommand struct {
	ID          string
	Description *string
	AmountCents *int64
	CategoryID  *string
	IsActive    *bool
}

// GenerateRecurringExpensesCommand holds parameters for generating expenses from recurring templates.
type GenerateRecurringExpensesCommand struct {
	HouseholdID string
	AsOfDate    time.Time
}

// GenerateRecurringExpensesOutput contains the result of expense generation.
type GenerateRecurringExpensesOutput struct {
	GeneratedCount int
	ExpenseIDs     []string
}

type CreateRecurringExpenseUseCase interface {
	Execute(ctx context.Context, input CreateRecurringExpenseInput) (CreateRecurringExpenseOutput, error)
}

type GetRecurringExpenseUseCase interface {
	Execute(ctx context.Context, id string) (recurringexpense.RecurringExpense, error)
}

type ListRecurringExpensesUseCase interface {
	Execute(ctx context.Context, query ListRecurringExpensesQuery) ([]recurringexpense.RecurringExpense, error)
}

type UpdateRecurringExpenseUseCase interface {
	Execute(ctx context.Context, cmd UpdateRecurringExpenseCommand) (recurringexpense.RecurringExpense, error)
}

type DeleteRecurringExpenseUseCase interface {
	Execute(ctx context.Context, id string) error
}

type GenerateRecurringExpensesUseCase interface {
	Execute(ctx context.Context, cmd GenerateRecurringExpensesCommand) (GenerateRecurringExpensesOutput, error)
}
