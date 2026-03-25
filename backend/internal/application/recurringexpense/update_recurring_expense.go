package recurringexpenseapp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type UpdateRecurringExpenseUseCase struct {
	repo         outbound.RecurringExpenseRepository
	categoryRepo outbound.CategoryRepository
}

func NewUpdateRecurringExpenseUseCase(
	repo outbound.RecurringExpenseRepository,
	categoryRepo outbound.CategoryRepository,
) UpdateRecurringExpenseUseCase {
	return UpdateRecurringExpenseUseCase{
		repo:         repo,
		categoryRepo: categoryRepo,
	}
}

func (u UpdateRecurringExpenseUseCase) Execute(ctx context.Context, cmd inbound.UpdateRecurringExpenseCommand) (recurringexpense.RecurringExpense, error) {
	re, err := u.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return recurringexpense.RecurringExpense{}, fmt.Errorf("update recurring expense: %w", err)
	}

	categoryID := cmd.CategoryID
	if categoryID != nil {
		resolved, err := u.resolveCategoryID(ctx, re.HouseholdID(), *categoryID)
		if err != nil {
			return recurringexpense.RecurringExpense{}, fmt.Errorf("update recurring expense: resolve category: %w", err)
		}
		categoryID = &resolved
	}

	if err := re.Patch(cmd.Description, cmd.AmountCents, categoryID, cmd.IsActive); err != nil {
		return recurringexpense.RecurringExpense{}, fmt.Errorf("update recurring expense: %w", err)
	}

	if err := u.repo.Update(ctx, re); err != nil {
		return recurringexpense.RecurringExpense{}, fmt.Errorf("update recurring expense: %w", err)
	}

	slog.InfoContext(ctx, "update recurring expense", "recurring_expense_id", string(re.ID()))
	return re, nil
}

func (u UpdateRecurringExpenseUseCase) resolveCategoryID(ctx context.Context, householdID, input string) (string, error) {
	input = strings.TrimSpace(input)

	// If it's a UUID, it's likely already a valid Category ID.
	if _, err := uuid.Parse(input); err == nil {
		return input, nil
	}

	// If it's empty or doesn't look like a UUID, we treat it as a slug.
	slug := input
	if slug == "" {
		slug = "other"
	}

	c, err := u.categoryRepo.FindBySlug(ctx, householdID, slug)
	if err != nil {
		// If the specific slug fails, try fallback to "other".
		if slug != "other" {
			c, err = u.categoryRepo.FindBySlug(ctx, householdID, "other")
		}
		if err != nil {
			return "", fmt.Errorf("failed to resolve category: %w", err)
		}
	}

	return string(c.ID()), nil
}

var _ inbound.UpdateRecurringExpenseUseCase = UpdateRecurringExpenseUseCase{}

var _ inbound.UpdateRecurringExpenseUseCase = UpdateRecurringExpenseUseCase{}
