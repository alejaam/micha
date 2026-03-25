package recurringexpenseapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type GetRecurringExpenseUseCase struct {
	repo outbound.RecurringExpenseRepository
}

func NewGetRecurringExpenseUseCase(repo outbound.RecurringExpenseRepository) GetRecurringExpenseUseCase {
	return GetRecurringExpenseUseCase{repo: repo}
}

func (u GetRecurringExpenseUseCase) Execute(ctx context.Context, id string) (recurringexpense.RecurringExpense, error) {
	re, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return recurringexpense.RecurringExpense{}, fmt.Errorf("get recurring expense: %w", err)
	}
	return re, nil
}

var _ inbound.GetRecurringExpenseUseCase = GetRecurringExpenseUseCase{}
