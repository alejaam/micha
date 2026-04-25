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
	ListByPeriod(ctx context.Context, periodID string) ([]expense.Expense, error)
	ListByHouseholdAndPeriod(ctx context.Context, householdID string, from, to time.Time) ([]expense.Expense, error)
	// SumPersonalByMemberAndPeriod returns personal outflow for one member in [from, to),
	// counting non-MSI expenses by creation date and MSI by installment month.
	SumPersonalByMemberAndPeriod(ctx context.Context, householdID, memberID string, from, to time.Time) (int64, error)
	Update(ctx context.Context, e expense.Expense) error
}
