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

var _ inbound.GetExpenseUseCase = GetExpenseUseCase{}

// GetExpenseUseCase retrieves a single non-deleted expense by ID.
type GetExpenseUseCase struct {
	repo       outbound.ExpenseRepository
	memberRepo outbound.MemberRepository
}

// NewGetExpenseUseCase constructs a GetExpenseUseCase.
func NewGetExpenseUseCase(repo outbound.ExpenseRepository, memberRepo outbound.MemberRepository) GetExpenseUseCase {
	return GetExpenseUseCase{repo: repo, memberRepo: memberRepo}
}

func (u GetExpenseUseCase) Execute(ctx context.Context, input inbound.GetExpenseInput) (expense.Expense, error) {
	e, err := u.repo.FindByID(ctx, input.ID)
	if err != nil {
		return expense.Expense{}, fmt.Errorf("get expense: %w", err)
	}

	if e.DeletedAt() != nil {
		slog.InfoContext(ctx, "get expense: already deleted", "expense_id", input.ID)
		return expense.Expense{}, shared.ErrNotFound
	}

	if _, err := u.memberRepo.FindByUserID(ctx, e.HouseholdID(), input.CurrentUserID); err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return expense.Expense{}, shared.ErrForbidden
		}
		return expense.Expense{}, fmt.Errorf("get expense: authorize member: %w", err)
	}

	return e, nil
}
