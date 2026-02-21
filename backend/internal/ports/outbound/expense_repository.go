package outbound

import (
	"context"

	"micha/backend/internal/domain/expense"
)

type ExpenseRepository interface {
	Save(ctx context.Context, e expense.Expense) error
}
