package expenseapp

import (
	"context"

	"micha/backend/internal/domain/expense"
)

// ExpenseRepository defines the persistence contract required by expense use cases.
type ExpenseRepository interface {
	Save(ctx context.Context, e expense.Expense) error
	FindByID(ctx context.Context, id string) (expense.Expense, error)
	List(ctx context.Context, householdID string, limit, offset int) ([]expense.Expense, error)
	Update(ctx context.Context, e expense.Expense) error
}

// RegisterExpenseInput contains required data to register an expense.
type RegisterExpenseInput struct {
	HouseholdID string
	AmountCents int64
	Description string
}

// RegisterExpenseOutput contains created expense identifiers.
type RegisterExpenseOutput struct {
	ExpenseID string
}

// ListExpensesQuery holds the parameters for listing expenses.
type ListExpensesQuery struct {
	HouseholdID string
	Limit       int
	Offset      int
}

// PatchExpenseCommand holds the partial-update fields for an expense.
// Only non-nil fields are applied.
type PatchExpenseCommand struct {
	ID          string
	Description *string
	AmountCents *int64
}
