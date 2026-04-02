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
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type RegisterExpenseUseCase struct {
	repo            outbound.ExpenseRepository
	householdRepo   outbound.HouseholdRepository
	memberRepo      outbound.MemberRepository
	cardRepo        outbound.CardRepository
	categoryRepo    outbound.CategoryRepository
	installmentRepo outbound.InstallmentRepository
	idGenerator     appshared.IDGenerator
	now             func() time.Time
}

func NewRegisterExpenseUseCase(
	repo outbound.ExpenseRepository,
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	cardRepo outbound.CardRepository,
	categoryRepo outbound.CategoryRepository,
	installmentRepo outbound.InstallmentRepository,
	idGenerator appshared.IDGenerator,
) RegisterExpenseUseCase {
	return RegisterExpenseUseCase{
		repo:            repo,
		householdRepo:   householdRepo,
		memberRepo:      memberRepo,
		cardRepo:        cardRepo,
		categoryRepo:    categoryRepo,
		installmentRepo: installmentRepo,
		idGenerator:     idGenerator,
		now:             time.Now,
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

	// Requirement: Pending members cannot register expenses.
	if m.IsPending() {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", shared.ErrForbidden)
	}

	// NEW: Validate session authorization — only the member can register their own expenses,
	// UNLESS the current user is the admin (first/creator member of the household).
	if err := u.validateSessionAuthorization(ctx, input.HouseholdID, input.PaidByMemberID, input.CurrentUserID); err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	categoryID, err := u.resolveCategoryID(ctx, input.HouseholdID, input.CategoryID)
	if err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: resolve category: %w", err)
	}

	cardID, cardName, err := u.resolveCardDetails(ctx, input)
	if err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: resolve card: %w", err)
	}

	now := u.now()
	e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:                expense.ID(u.idGenerator.NewID()),
		HouseholdID:       input.HouseholdID,
		PaidByMemberID:    input.PaidByMemberID,
		AmountCents:       input.AmountCents,
		Description:       input.Description,
		IsShared:          input.IsShared,
		Currency:          input.Currency,
		PaymentMethod:     expense.PaymentMethod(input.PaymentMethod),
		ExpenseType:       expense.ExpenseType(input.ExpenseType),
		CardID:            cardID,
		CardName:          cardName,
		CategoryID:        categoryID,
		TotalInstallments: input.TotalInstallments,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	if err := u.repo.Save(ctx, e); err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	// Requirement: Generate installments for MSI expenses.
	if e.ExpenseType() == expense.ExpenseTypeMSI {
		installments := u.generateInstallments(e)
		if err := u.installmentRepo.SaveAll(ctx, installments); err != nil {
			return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: save installments: %w", err)
		}
	}

	slog.InfoContext(ctx, "register expense", "expense_id", string(e.ID()))
	return inbound.RegisterExpenseOutput{ExpenseID: string(e.ID())}, nil
}

func (u RegisterExpenseUseCase) generateInstallments(root expense.Expense) []installment.Installment {
	attrs := root.Attributes()
	count := attrs.TotalInstallments
	total := attrs.AmountCents

	base := total / int64(count)
	remainder := total % int64(count)

	installments := make([]installment.Installment, count)
	for i := 0; i < count; i++ {
		amount := base
		if int64(i) < remainder {
			amount++
		}

		// Monthly increments from root expense date.
		startDate := attrs.CreatedAt.AddDate(0, i, 0)

		inst, _ := installment.New(
			installment.ID(u.idGenerator.NewID()),
			string(attrs.ID),
			attrs.PaidByMemberID,
			startDate,
			amount,
			total,
			count,
			i+1,
			attrs.CreatedAt,
		)
		installments[i] = inst
	}
	return installments
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

func (u RegisterExpenseUseCase) resolveCardDetails(ctx context.Context, input inbound.RegisterExpenseInput) (string, string, error) {
	if expense.PaymentMethod(input.PaymentMethod) != expense.PaymentMethodCard {
		return "", "", nil
	}

	cardID := strings.TrimSpace(input.CardID)
	if cardID == "" {
		return "", strings.TrimSpace(input.CardName), nil
	}

	c, err := u.cardRepo.FindByID(ctx, cardID)
	if err != nil {
		return "", "", err
	}
	if c.HouseholdID() != input.HouseholdID {
		return "", "", shared.ErrNotFound
	}

	return string(c.ID()), c.CardName(), nil
}

// validateSessionAuthorization verifies that the current user is authorized to register an expense
// for the specified member. Rules:
// 1. A user can always register an expense for their own member (m.UserID() == currentUserID)
// 2. A user can register expenses for other members if they are the admin (first member linked to the household creator)
func (u RegisterExpenseUseCase) validateSessionAuthorization(ctx context.Context, householdID, paidByMemberID, currentUserID string) error {
	if strings.TrimSpace(currentUserID) == "" {
		return shared.ErrForbidden // No user context
	}

	// Get the member paying for the expense
	paidByMember, err := u.memberRepo.FindByID(ctx, paidByMemberID)
	if err != nil {
		return err
	}

	// Rule 1: User can register their own expenses
	if paidByMember.UserID() == currentUserID {
		return nil
	}

	// Rule 2: User must be the admin (first member of household with a user_id)
	members, err := u.memberRepo.ListAllByHousehold(ctx, householdID)
	if err != nil {
		return err
	}

	// Find the admin: first member with a non-empty user_id
	for _, m := range members {
		if strings.TrimSpace(m.UserID()) != "" {
			// Found the first member with a user_id; they are the admin
			if m.UserID() == currentUserID {
				return nil // Current user is the admin, allowed
			}
			break // Stop checking; this is the admin and it's not the current user
		}
	}

	return shared.ErrForbidden
}

var _ inbound.RegisterExpenseUseCase = RegisterExpenseUseCase{}
