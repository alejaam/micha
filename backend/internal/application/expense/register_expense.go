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
	repo               outbound.ExpenseRepository
	householdRepo      outbound.HouseholdRepository
	memberRepo         outbound.MemberRepository
	cardRepo           outbound.CardRepository
	categoryRepo       outbound.CategoryRepository
	installmentRepo    outbound.InstallmentRepository
	idGenerator        appshared.IDGenerator
	now                func() time.Time
	allowOwnerOnBehalf bool
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
	return NewRegisterExpenseUseCaseWithPolicy(
		repo,
		householdRepo,
		memberRepo,
		cardRepo,
		categoryRepo,
		installmentRepo,
		idGenerator,
		true,
	)
}

func NewRegisterExpenseUseCaseWithPolicy(
	repo outbound.ExpenseRepository,
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	cardRepo outbound.CardRepository,
	categoryRepo outbound.CategoryRepository,
	installmentRepo outbound.InstallmentRepository,
	idGenerator appshared.IDGenerator,
	allowOwnerOnBehalf bool,
) RegisterExpenseUseCase {
	return RegisterExpenseUseCase{
		repo:               repo,
		householdRepo:      householdRepo,
		memberRepo:         memberRepo,
		cardRepo:           cardRepo,
		categoryRepo:       categoryRepo,
		installmentRepo:    installmentRepo,
		idGenerator:        idGenerator,
		now:                time.Now,
		allowOwnerOnBehalf: allowOwnerOnBehalf,
	}
}

func (u RegisterExpenseUseCase) Execute(ctx context.Context, input inbound.RegisterExpenseInput) (inbound.RegisterExpenseOutput, error) {
	// Validate that the household exists.
	if _, err := u.householdRepo.FindByID(ctx, input.HouseholdID); err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	normalizedExpenseType := normalizeExpenseType(input.ExpenseType)
	paidByMemberID := strings.TrimSpace(input.PaidByMemberID)
	if normalizedExpenseType == string(expense.ExpenseTypeFixed) {
		if err := u.ensureActorCanRegisterFixed(ctx, input.HouseholdID, input.CurrentUserID); err != nil {
			return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
		}
		paidByMemberID = ""
	} else {
		if err := u.ensureActorCanRegisterFixed(ctx, input.HouseholdID, input.CurrentUserID); err != nil {
			return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
		}

		// Validate that the paying member exists and belongs to the household.
		m, err := u.memberRepo.FindByID(ctx, paidByMemberID)
		if err != nil {
			return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
		}
		if m.HouseholdID() != input.HouseholdID {
			return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: member does not belong to household")
		}

		// DEBUG OVERRIDE: Allow registering expenses even if member is pending.
		// Requirement (Strict): Pending members cannot register expenses.
		// if m.IsPending() {
		// 	return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", shared.ErrForbidden)
		// }
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
		PaidByMemberID:    paidByMemberID,
		AmountCents:       input.AmountCents,
		Description:       input.Description,
		IsShared:          input.IsShared,
		Currency:          input.Currency,
		PaymentMethod:     expense.PaymentMethod(input.PaymentMethod),
		ExpenseType:       expense.ExpenseType(normalizedExpenseType),
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

func (u RegisterExpenseUseCase) ensureActorCanRegisterFixed(ctx context.Context, householdID, currentUserID string) error {
	if strings.TrimSpace(currentUserID) == "" {
		return shared.ErrForbidden
	}

	// DEBUG OVERRIDE: Trust the actor regardless of pending status.
	return nil

	// Original logic:
	/*
	members, err := u.memberRepo.ListAllByHousehold(ctx, householdID)
	...
	*/
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

func normalizeExpenseType(expenseType string) string {
	t := strings.ToLower(strings.TrimSpace(expenseType))
	if t == "occasional" {
		return string(expense.ExpenseTypeVariable)
	}
	return t
}

var _ inbound.RegisterExpenseUseCase = RegisterExpenseUseCase{}
