package outbound

import (
	"context"
	"time"

	"micha/backend/internal/domain/expense"
)

// ExpenseRepository defines the persistence contract required by expense use cases.
type ExpenseRepository interface {
	Save(ctx context.Context, e expense.Expense) error
	FindByID(ctx context.Context, id string) (expense.Expense, error)
	List(ctx context.Context, householdID string, limit, offset int) ([]expense.Expense, error)
	ListByHouseholdAndPeriod(ctx context.Context, householdID string, from, to time.Time) ([]expense.Expense, error)
	Update(ctx context.Context, e expense.Expense) error
}
