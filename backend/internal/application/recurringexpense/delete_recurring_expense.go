package recurringexpenseapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type DeleteRecurringExpenseUseCase struct {
	repo       outbound.RecurringExpenseRepository
	memberRepo outbound.MemberRepository
}

func NewDeleteRecurringExpenseUseCase(repo outbound.RecurringExpenseRepository, memberRepo outbound.MemberRepository) DeleteRecurringExpenseUseCase {
	return DeleteRecurringExpenseUseCase{repo: repo, memberRepo: memberRepo}
}

func (u DeleteRecurringExpenseUseCase) Execute(ctx context.Context, input inbound.DeleteRecurringExpenseInput) error {
	re, err := u.repo.FindByID(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("delete recurring expense: %w", err)
	}
	if err := authorizeActiveHouseholdMember(ctx, u.memberRepo, re.HouseholdID(), input.CurrentUserID); err != nil {
		return fmt.Errorf("delete recurring expense: %w", err)
	}

	if err := re.SoftDelete(); err != nil {
		return fmt.Errorf("delete recurring expense: %w", err)
	}

	if err := u.repo.Update(ctx, re); err != nil {
		return fmt.Errorf("delete recurring expense: %w", err)
	}

	slog.InfoContext(ctx, "delete recurring expense", "recurring_expense_id", input.ID)
	return nil
}

var _ inbound.DeleteRecurringExpenseUseCase = DeleteRecurringExpenseUseCase{}
