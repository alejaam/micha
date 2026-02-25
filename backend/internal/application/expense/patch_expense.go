package expenseapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/expense"
)

// PatchExpenseUseCase applies a partial update to an expense.
type PatchExpenseUseCase struct {
	repo ExpenseRepository
}

// NewPatchExpenseUseCase constructs a PatchExpenseUseCase.
func NewPatchExpenseUseCase(repo ExpenseRepository) PatchExpenseUseCase {
	return PatchExpenseUseCase{repo: repo}
}

func (u PatchExpenseUseCase) Execute(ctx context.Context, cmd PatchExpenseCommand) (expense.Expense, error) {
	e, err := u.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return expense.Expense{}, fmt.Errorf("patch expense: %w", err)
	}

	if err := e.Patch(cmd.Description, cmd.AmountCents); err != nil {
		return expense.Expense{}, fmt.Errorf("patch expense: %w", err)
	}

	if err := u.repo.Update(ctx, e); err != nil {
		return expense.Expense{}, fmt.Errorf("patch expense: %w", err)
	}

	slog.InfoContext(ctx, "patch expense", "expense_id", cmd.ID)
	return e, nil
}
