package expenseapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

// GetExpenseUseCase retrieves a single non-deleted expense by ID.
type GetExpenseUseCase struct {
	repo outbound.ExpenseRepository
}

// NewGetExpenseUseCase constructs a GetExpenseUseCase.
func NewGetExpenseUseCase(repo outbound.ExpenseRepository) GetExpenseUseCase {
	return GetExpenseUseCase{repo: repo}
}

func (u GetExpenseUseCase) Execute(ctx context.Context, id string) (expense.Expense, error) {
	e, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return expense.Expense{}, fmt.Errorf("get expense: %w", err)
	}

	if e.DeletedAt() != nil {
		slog.InfoContext(ctx, "get expense: already deleted", "expense_id", id)
		return expense.Expense{}, shared.ErrNotFound
	}

	return e, nil
}
