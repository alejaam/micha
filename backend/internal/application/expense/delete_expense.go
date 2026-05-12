package expenseapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

var _ inbound.DeleteExpenseUseCase = DeleteExpenseUseCase{}

// DeleteExpenseUseCase soft-deletes an expense by ID.
// For MSI expenses, it cascades to delete related installments first.
type DeleteExpenseUseCase struct {
	repo            outbound.ExpenseRepository
	installmentRepo outbound.InstallmentRepository
}

// NewDeleteExpenseUseCase constructs a DeleteExpenseUseCase.
func NewDeleteExpenseUseCase(repo outbound.ExpenseRepository, installmentRepo outbound.InstallmentRepository) DeleteExpenseUseCase {
	return DeleteExpenseUseCase{repo: repo, installmentRepo: installmentRepo}
}

func (u DeleteExpenseUseCase) Execute(ctx context.Context, id string) error {
	e, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	// Cascade delete installments for MSI expenses before soft-deleting the parent.
	if e.ExpenseType() == expense.ExpenseTypeMSI {
		if err := u.installmentRepo.DeleteByExpense(ctx, string(e.ID())); err != nil {
			return fmt.Errorf("delete expense: cascade delete installments: %w", err)
		}
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
