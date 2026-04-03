package outbound

import (
	"context"
	"time"

	"micha/backend/internal/domain/installment"
)

// InstallmentRepository defines the persistence contract required for installments.
type InstallmentRepository interface {
	Save(ctx context.Context, inst installment.Installment) error
	SaveAll(ctx context.Context, insts []installment.Installment) error
	ListByExpense(ctx context.Context, expenseID string) ([]installment.Installment, error)
	ListByHouseholdAndPeriod(ctx context.Context, householdID string, from, to time.Time) ([]installment.Installment, error)
	DeleteByExpense(ctx context.Context, expenseID string) error
}
