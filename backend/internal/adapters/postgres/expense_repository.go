package postgres

import (
	"context"

	"micha/backend/internal/domain/expense"
)

type ExpenseRepository struct{}

func NewExpenseRepository() ExpenseRepository {
	return ExpenseRepository{}
}

func (r ExpenseRepository) Save(_ context.Context, _ expense.Expense) error {
	return nil
}
