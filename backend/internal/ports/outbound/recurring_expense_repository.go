package outbound

import (
	"context"
	"time"

	"micha/backend/internal/domain/recurringexpense"
)

// RecurringExpenseRepository defines the persistence contract required by recurring expense use cases.
type RecurringExpenseRepository interface {
	Save(ctx context.Context, r recurringexpense.RecurringExpense) error
	FindByID(ctx context.Context, id string) (recurringexpense.RecurringExpense, error)
	List(ctx context.Context, householdID string, limit, offset int) ([]recurringexpense.RecurringExpense, error)
	ListDueForGeneration(ctx context.Context, asOfDate time.Time) ([]recurringexpense.RecurringExpense, error)
	Update(ctx context.Context, r recurringexpense.RecurringExpense) error
}
