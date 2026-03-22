package expenseapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/application/shared"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ListExpensesUseCase lists non-deleted expenses for a household with pagination.
type ListExpensesUseCase struct {
	repo outbound.ExpenseRepository
}

// NewListExpensesUseCase constructs a ListExpensesUseCase.
func NewListExpensesUseCase(repo outbound.ExpenseRepository) ListExpensesUseCase {
	return ListExpensesUseCase{repo: repo}
}

func (u ListExpensesUseCase) Execute(ctx context.Context, query inbound.ListExpensesQuery) ([]expense.Expense, error) {
	if query.HouseholdID == "" {
		return nil, fmt.Errorf("list expenses: household_id is required")
	}

	limit := query.Limit
	if limit <= 0 {
		limit = appshared.DefaultLimit
	}
	if limit > appshared.MaxLimit {
		limit = appshared.MaxLimit
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	expenses, err := u.repo.List(ctx, query.HouseholdID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}

	slog.InfoContext(ctx, "list expenses", "household_id", query.HouseholdID, "limit", limit, "offset", offset, "count", len(expenses))
	return expenses, nil
}

var _ inbound.ListExpensesUseCase = ListExpensesUseCase{}
