package recurringexpenseapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type ListRecurringExpensesUseCase struct {
	repo outbound.RecurringExpenseRepository
}

func NewListRecurringExpensesUseCase(repo outbound.RecurringExpenseRepository) ListRecurringExpensesUseCase {
	return ListRecurringExpensesUseCase{repo: repo}
}

func (u ListRecurringExpensesUseCase) Execute(ctx context.Context, query inbound.ListRecurringExpensesQuery) ([]recurringexpense.RecurringExpense, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}

	recurringExpenses, err := u.repo.List(ctx, query.HouseholdID, limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("list recurring expenses: %w", err)
	}

	return recurringExpenses, nil
}

var _ inbound.ListRecurringExpensesUseCase = ListRecurringExpensesUseCase{}
