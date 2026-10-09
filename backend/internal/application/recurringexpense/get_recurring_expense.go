package recurringexpenseapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type GetRecurringExpenseUseCase struct {
	repo       outbound.RecurringExpenseRepository
	memberRepo outbound.MemberRepository
}

func NewGetRecurringExpenseUseCase(repo outbound.RecurringExpenseRepository, memberRepo outbound.MemberRepository) GetRecurringExpenseUseCase {
	return GetRecurringExpenseUseCase{repo: repo, memberRepo: memberRepo}
}

func (u GetRecurringExpenseUseCase) Execute(ctx context.Context, input inbound.GetRecurringExpenseInput) (recurringexpense.RecurringExpense, error) {
	re, err := u.repo.FindByID(ctx, input.ID)
	if err != nil {
		return recurringexpense.RecurringExpense{}, fmt.Errorf("get recurring expense: %w", err)
	}
	if err := authorizeActiveHouseholdMember(ctx, u.memberRepo, re.HouseholdID(), input.CurrentUserID); err != nil {
		return recurringexpense.RecurringExpense{}, fmt.Errorf("get recurring expense: %w", err)
	}
	return re, nil
}

var _ inbound.GetRecurringExpenseUseCase = GetRecurringExpenseUseCase{}
