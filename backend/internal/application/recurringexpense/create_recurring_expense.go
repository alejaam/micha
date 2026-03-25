package recurringexpenseapp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type CreateRecurringExpenseUseCase struct {
	repo          outbound.RecurringExpenseRepository
	householdRepo outbound.HouseholdRepository
	memberRepo    outbound.MemberRepository
	categoryRepo  outbound.CategoryRepository
	idGenerator   appshared.IDGenerator
	now           func() time.Time
}

func NewCreateRecurringExpenseUseCase(
	repo outbound.RecurringExpenseRepository,
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	categoryRepo outbound.CategoryRepository,
	idGenerator appshared.IDGenerator,
) CreateRecurringExpenseUseCase {
	return CreateRecurringExpenseUseCase{
		repo:          repo,
		householdRepo: householdRepo,
		memberRepo:    memberRepo,
		categoryRepo:  categoryRepo,
		idGenerator:   idGenerator,
		now:           time.Now,
	}
}

func (u CreateRecurringExpenseUseCase) Execute(ctx context.Context, input inbound.CreateRecurringExpenseInput) (inbound.CreateRecurringExpenseOutput, error) {
	// Validate that the household exists.
	if _, err := u.householdRepo.FindByID(ctx, input.HouseholdID); err != nil {
		return inbound.CreateRecurringExpenseOutput{}, fmt.Errorf("create recurring expense: %w", err)
	}

	// Validate that the paying member exists and belongs to the household.
	m, err := u.memberRepo.FindByID(ctx, input.PaidByMemberID)
	if err != nil {
		return inbound.CreateRecurringExpenseOutput{}, fmt.Errorf("create recurring expense: %w", err)
	}
	if m.HouseholdID() != input.HouseholdID {
		return inbound.CreateRecurringExpenseOutput{}, fmt.Errorf("create recurring expense: member does not belong to household")
	}

	categoryID, err := u.resolveCategoryID(ctx, input.HouseholdID, input.CategoryID)
	if err != nil {
		return inbound.CreateRecurringExpenseOutput{}, fmt.Errorf("create recurring expense: resolve category: %w", err)
	}

	now := u.now()
	re, err := recurringexpense.New(
		recurringexpense.ID(u.idGenerator.NewID()),
		input.HouseholdID,
		input.PaidByMemberID,
		input.AmountCents,
		input.Description,
		categoryID,
		expense.ExpenseType(input.ExpenseType),
		recurringexpense.RecurrencePattern(input.RecurrencePattern),
		input.StartDate,
		now,
	)
	if err != nil {
		return inbound.CreateRecurringExpenseOutput{}, fmt.Errorf("create recurring expense: %w", err)
	}

	// If end date was provided, we need to reconstruct with it
	if input.EndDate != nil {
		attrs := re.Attributes()
		attrs.EndDate = input.EndDate
		re, err = recurringexpense.NewFromAttributes(attrs)
		if err != nil {
			return inbound.CreateRecurringExpenseOutput{}, fmt.Errorf("create recurring expense: %w", err)
		}
	}

	if err := u.repo.Save(ctx, re); err != nil {
		return inbound.CreateRecurringExpenseOutput{}, fmt.Errorf("create recurring expense: %w", err)
	}

	slog.InfoContext(ctx, "create recurring expense", "recurring_expense_id", string(re.ID()))
	return inbound.CreateRecurringExpenseOutput{RecurringExpenseID: string(re.ID())}, nil
}

func (u CreateRecurringExpenseUseCase) resolveCategoryID(ctx context.Context, householdID, input string) (string, error) {
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

var _ inbound.CreateRecurringExpenseUseCase = CreateRecurringExpenseUseCase{}

var _ inbound.CreateRecurringExpenseUseCase = CreateRecurringExpenseUseCase{}
