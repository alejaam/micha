package inbound

import (
	"context"

	"micha/backend/internal/domain/expense"
)

// RegisterExpenseInput contains required data to register an expense.
type RegisterExpenseInput struct {
	HouseholdID       string
	PaidByMemberID    string
	CurrentUserID     string // The user ID of the request caller (from JWT)
	AmountCents       int64
	Description       string
	IsShared          bool
	Currency          string
	PaymentMethod     string
	ExpenseType       string
	CardID            string
	CardName          string
	CategoryID        string
	TotalInstallments int
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

type RegisterExpenseUseCase interface {
	Execute(ctx context.Context, input RegisterExpenseInput) (RegisterExpenseOutput, error)
}

type GetExpenseUseCase interface {
	Execute(ctx context.Context, id string) (expense.Expense, error)
}

type ListExpensesUseCase interface {
	Execute(ctx context.Context, query ListExpensesQuery) ([]expense.Expense, error)
}

type PatchExpenseUseCase interface {
	Execute(ctx context.Context, cmd PatchExpenseCommand) (expense.Expense, error)
}

type DeleteExpenseUseCase interface {
	Execute(ctx context.Context, id string) error
}
