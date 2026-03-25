package expenseapp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type RegisterExpenseUseCase struct {
	repo          outbound.ExpenseRepository
	householdRepo outbound.HouseholdRepository
	memberRepo    outbound.MemberRepository
	categoryRepo  outbound.CategoryRepository
	idGenerator   appshared.IDGenerator
	now           func() time.Time
}

func NewRegisterExpenseUseCase(
	repo outbound.ExpenseRepository,
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	categoryRepo outbound.CategoryRepository,
	idGenerator appshared.IDGenerator,
) RegisterExpenseUseCase {
	return RegisterExpenseUseCase{
		repo:          repo,
		householdRepo: householdRepo,
		memberRepo:    memberRepo,
		categoryRepo:  categoryRepo,
		idGenerator:   idGenerator,
		now:           time.Now,
	}
}

func (u RegisterExpenseUseCase) Execute(ctx context.Context, input inbound.RegisterExpenseInput) (inbound.RegisterExpenseOutput, error) {
	// Validate that the household exists.
	if _, err := u.householdRepo.FindByID(ctx, input.HouseholdID); err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	// Validate that the paying member exists and belongs to the household.
	m, err := u.memberRepo.FindByID(ctx, input.PaidByMemberID)
	if err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}
	if m.HouseholdID() != input.HouseholdID {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: member does not belong to household")
	}

	categoryID, err := u.resolveCategoryID(ctx, input.HouseholdID, input.CategoryID)
	if err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: resolve category: %w", err)
	}

	now := u.now()
	e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID(u.idGenerator.NewID()),
		HouseholdID:    input.HouseholdID,
		PaidByMemberID: input.PaidByMemberID,
		AmountCents:    input.AmountCents,
		Description:    input.Description,
		IsShared:       input.IsShared,
		Currency:       input.Currency,
		PaymentMethod:  expense.PaymentMethod(input.PaymentMethod),
		ExpenseType:    expense.ExpenseType(input.ExpenseType),
		CardName:       input.CardName,
		CategoryID:     categoryID,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	if err := u.repo.Save(ctx, e); err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	slog.InfoContext(ctx, "register expense", "expense_id", string(e.ID()))
	return inbound.RegisterExpenseOutput{ExpenseID: string(e.ID())}, nil
}

func (u RegisterExpenseUseCase) resolveCategoryID(ctx context.Context, householdID, input string) (string, error) {
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

var _ inbound.RegisterExpenseUseCase = RegisterExpenseUseCase{}
