package recurringexpenseapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type DeleteRecurringExpenseUseCase struct {
	repo outbound.RecurringExpenseRepository
}

func NewDeleteRecurringExpenseUseCase(repo outbound.RecurringExpenseRepository) DeleteRecurringExpenseUseCase {
	return DeleteRecurringExpenseUseCase{repo: repo}
}

func (u DeleteRecurringExpenseUseCase) Execute(ctx context.Context, id string) error {
	re, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete recurring expense: %w", err)
	}

	if err := re.SoftDelete(); err != nil {
		return fmt.Errorf("delete recurring expense: %w", err)
	}

	if err := u.repo.Update(ctx, re); err != nil {
		return fmt.Errorf("delete recurring expense: %w", err)
	}

	slog.InfoContext(ctx, "delete recurring expense", "recurring_expense_id", string(re.ID()))
	return nil
}

var _ inbound.DeleteRecurringExpenseUseCase = DeleteRecurringExpenseUseCase{}
