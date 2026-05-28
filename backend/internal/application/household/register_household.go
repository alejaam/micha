package householdapp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"

	periodapp "micha/backend/internal/application/period"
)

var _ inbound.RegisterHouseholdUseCase = RegisterHouseholdUseCase{}

// defaultCategoryDefs maps each default slug to its display name.
var defaultCategoryDefs = []struct {
	name string
	slug string
}{
	{"Rent", "rent"},
	{"Auto", "auto"},
	{"Streaming", "streaming"},
	{"Food", "food"},
	{"Personal", "personal"},
	{"Savings", "savings"},
	{"Other", "other"},
}

// RegisterHouseholdUseCase creates a new household, seeds default categories,
// registers the current user as the owner member (with salary), and auto-creates
// the first period — all in a single transaction.
type RegisterHouseholdUseCase struct {
	repo         outbound.HouseholdRepository
	categoryRepo outbound.CategoryRepository
	memberRepo   outbound.MemberRepository
	userRepo     outbound.UserRepository
	periodRepo   outbound.PeriodRepository
	txManager    outbound.TransactionManager
	idGenerator  appshared.IDGenerator
	Now          func() time.Time
}

// NewRegisterHouseholdUseCase constructs RegisterHouseholdUseCase.
func NewRegisterHouseholdUseCase(
	repo outbound.HouseholdRepository,
	categoryRepo outbound.CategoryRepository,
	memberRepo outbound.MemberRepository,
	userRepo outbound.UserRepository,
	periodRepo outbound.PeriodRepository,
	txManager outbound.TransactionManager,
	idGenerator appshared.IDGenerator,
) RegisterHouseholdUseCase {
	return RegisterHouseholdUseCase{
		repo:         repo,
		categoryRepo: categoryRepo,
		memberRepo:   memberRepo,
		userRepo:     userRepo,
		periodRepo:   periodRepo,
		txManager:    txManager,
		idGenerator:  idGenerator,
		Now:          appshared.Now,
	}
}

// Execute creates a household, stores it, seeds default categories,
// registers the owner member (with salary), and creates the first period atomically.
func (u RegisterHouseholdUseCase) Execute(ctx context.Context, input inbound.RegisterHouseholdInput) (inbound.RegisterHouseholdOutput, error) {
	if input.OwnerSalaryCents < 0 {
		return inbound.RegisterHouseholdOutput{}, fmt.Errorf("register household: %w", shared.ErrInvalidMoney)
	}

	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(u.idGenerator.NewID()),
		Name:            input.Name,
		OwnerID:         input.CurrentUserID,
		SettlementMode:  input.SettlementMode,
		Currency:        input.Currency,
		ClosingDay:      input.ClosingDay,
		PeriodFrequency: input.PeriodFrequency,
		CreatedAt:       u.Now(),
	})
	if err != nil {
		return inbound.RegisterHouseholdOutput{}, fmt.Errorf("register household: %w", err)
	}

	var output inbound.RegisterHouseholdOutput

	if err := u.txManager.Run(ctx, func(txCtx context.Context) error {
		// 1. Save household
		if err := u.repo.Save(txCtx, h); err != nil {
			return fmt.Errorf("register household: %w", err)
		}

		householdID := string(h.ID())
		now := u.Now()

		// 2. Create owner member with salary
		memberID, err := u.createOwnerMember(txCtx, householdID, input.CurrentUserID, input.OwnerSalaryCents, now)
		if err != nil {
			return err
		}

		// 3. Create initial period using shared helper
		result, err := periodapp.CreateInitialPeriod(
			txCtx, u.periodRepo, u.idGenerator,
			householdID, input.ClosingDay, input.PeriodFrequency, now,
		)
		if err != nil {
			return fmt.Errorf("register household: %w", err)
		}

		// 4. Seed default categories
		u.seedDefaultCategories(txCtx, householdID, now)

		output = inbound.RegisterHouseholdOutput{
			HouseholdID: householdID,
			MemberID:    memberID,
			PeriodID:    result.PeriodID,
		}
		return nil
	}); err != nil {
		return inbound.RegisterHouseholdOutput{}, err
	}

	slog.InfoContext(ctx, "register household",
		"household_id", output.HouseholdID,
		"member_id", output.MemberID,
		"period_id", output.PeriodID,
		"default_categories_seeded", len(defaultCategoryDefs),
	)
	return output, nil
}

// createOwnerMember creates the first (owner) member for the household with the given salary.
func (u RegisterHouseholdUseCase) createOwnerMember(ctx context.Context, householdID, userID string, salaryCents int64, now time.Time) (string, error) {
	usr, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("register household: failed to find user: %w", err)
	}

	email := usr.Email()
	name := email
	if at := strings.Index(email, "@"); at > 0 {
		name = strings.ToUpper(email[:1]) + email[1:at]
	}

	memberID := u.idGenerator.NewID()
	m, err := member.NewWithUserID(
		member.ID(memberID),
		householdID,
		name,
		email,
		userID,
		salaryCents,
		now,
	)
	if err != nil {
		return "", fmt.Errorf("register household: failed to create owner member: %w", err)
	}

	if err := u.memberRepo.Save(ctx, m); err != nil {
		return "", fmt.Errorf("register household: failed to save owner member: %w", err)
	}

	return memberID, nil
}

// seedDefaultCategories creates the default category set for the household.
func (u RegisterHouseholdUseCase) seedDefaultCategories(ctx context.Context, householdID string, now time.Time) {
	for _, def := range defaultCategoryDefs {
		cat, catErr := category.New(u.idGenerator.NewID(), householdID, def.name, def.slug, now)
		if catErr != nil {
			slog.WarnContext(ctx, "failed to create default category", "slug", def.slug, "error", catErr)
			continue
		}
		if saveErr := u.categoryRepo.Save(ctx, cat); saveErr != nil {
			slog.WarnContext(ctx, "failed to save default category", "slug", def.slug, "error", saveErr)
		}
	}
}
