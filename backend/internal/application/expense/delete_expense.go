package expenseapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

var _ inbound.DeleteExpenseUseCase = DeleteExpenseUseCase{}

// DeleteExpenseUseCase soft-deletes an expense by ID.
// For MSI expenses, it cascades to delete related installments first.
type DeleteExpenseUseCase struct {
	repo            outbound.ExpenseRepository
	memberRepo      outbound.MemberRepository
	installmentRepo outbound.InstallmentRepository
}

// NewDeleteExpenseUseCase constructs a DeleteExpenseUseCase.
func NewDeleteExpenseUseCase(repo outbound.ExpenseRepository, memberRepo outbound.MemberRepository, installmentRepo outbound.InstallmentRepository) DeleteExpenseUseCase {
	return DeleteExpenseUseCase{repo: repo, memberRepo: memberRepo, installmentRepo: installmentRepo}
}

func (u DeleteExpenseUseCase) Execute(ctx context.Context, input inbound.DeleteExpenseInput) error {
	e, err := u.repo.FindByID(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("delete expense: %w", err)
	}

	if _, err := u.memberRepo.FindByUserID(ctx, e.HouseholdID(), input.CurrentUserID); err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return shared.ErrForbidden
		}
		return fmt.Errorf("delete expense: authorize member: %w", err)
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

	slog.InfoContext(ctx, "delete expense", "expense_id", input.ID)
	return nil
}
