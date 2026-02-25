package expenseapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/expense"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ListExpensesUseCase lists non-deleted expenses for a household with pagination.
type ListExpensesUseCase struct {
	repo ExpenseRepository
}

// NewListExpensesUseCase constructs a ListExpensesUseCase.
func NewListExpensesUseCase(repo ExpenseRepository) ListExpensesUseCase {
	return ListExpensesUseCase{repo: repo}
}

func (u ListExpensesUseCase) Execute(ctx context.Context, query ListExpensesQuery) ([]expense.Expense, error) {
	if query.HouseholdID == "" {
		return nil, fmt.Errorf("list expenses: household_id is required")
	}

	limit := query.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
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
