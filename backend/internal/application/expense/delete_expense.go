package expenseapp

import (
	"context"
	"fmt"
	"log/slog"
)

// DeleteExpenseUseCase soft-deletes an expense by ID.
type DeleteExpenseUseCase struct {
	repo ExpenseRepository
}

// NewDeleteExpenseUseCase constructs a DeleteExpenseUseCase.
func NewDeleteExpenseUseCase(repo ExpenseRepository) DeleteExpenseUseCase {
	return DeleteExpenseUseCase{repo: repo}
}

func (u DeleteExpenseUseCase) Execute(ctx context.Context, id string) error {
	e, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	if err := e.SoftDelete(); err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	if err := u.repo.Update(ctx, e); err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	slog.InfoContext(ctx, "delete expense", "expense_id", id)
	return nil
}
