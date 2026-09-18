package expenseapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ListExpensesUseCase lists non-deleted expenses for a household with pagination.
type ListExpensesUseCase struct {
	repo       outbound.ExpenseRepository
	memberRepo outbound.MemberRepository
	periodRepo outbound.PeriodRepository
}

// NewListExpensesUseCase constructs a ListExpensesUseCase.
func NewListExpensesUseCase(repo outbound.ExpenseRepository, memberRepo outbound.MemberRepository, periodRepo outbound.PeriodRepository) ListExpensesUseCase {
	return ListExpensesUseCase{repo: repo, memberRepo: memberRepo, periodRepo: periodRepo}
}

func (u ListExpensesUseCase) Execute(ctx context.Context, query inbound.ListExpensesQuery) ([]expense.Expense, error) {
	if query.HouseholdID == "" {
		return nil, fmt.Errorf("list expenses: household_id is required")
	}

	if _, err := u.memberRepo.FindByUserID(ctx, query.HouseholdID, query.CurrentUserID); err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, shared.ErrForbidden
		}
		return nil, fmt.Errorf("list expenses: authorize member: %w", err)
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

	var expenses []expense.Expense
	var err error
	if query.PeriodID != "" {
		p, err := u.periodRepo.GetByID(ctx, period.ID(query.PeriodID))
		if err != nil {
			return nil, fmt.Errorf("list expenses: get period: %w", err)
		}
		if p.HouseholdID() != query.HouseholdID {
			return nil, shared.ErrForbidden
		}
		expenses, err = u.repo.ListByPeriod(ctx, query.PeriodID)
	} else {
		expenses, err = u.repo.List(ctx, query.HouseholdID, limit, offset)
	}
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}

	slog.InfoContext(ctx, "list expenses", "household_id", query.HouseholdID, "period_id", query.PeriodID, "limit", limit, "offset", offset, "count", len(expenses))
	return expenses, nil
}

var _ inbound.ListExpensesUseCase = ListExpensesUseCase{}
