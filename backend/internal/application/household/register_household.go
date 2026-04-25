package householdapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
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

// RegisterHouseholdUseCase creates a new household and seeds default categories.
type RegisterHouseholdUseCase struct {
	repo         outbound.HouseholdRepository
	categoryRepo outbound.CategoryRepository
	idGenerator  appshared.IDGenerator
	now          func() time.Time
}

// NewRegisterHouseholdUseCase constructs RegisterHouseholdUseCase.
func NewRegisterHouseholdUseCase(repo outbound.HouseholdRepository, categoryRepo outbound.CategoryRepository, idGenerator appshared.IDGenerator) RegisterHouseholdUseCase {
	return RegisterHouseholdUseCase{repo: repo, categoryRepo: categoryRepo, idGenerator: idGenerator, now: time.Now}
}

// Execute creates a household, stores it, and seeds its default categories.
func (u RegisterHouseholdUseCase) Execute(ctx context.Context, input inbound.RegisterHouseholdInput) (inbound.RegisterHouseholdOutput, error) {
	h, err := household.New(
		household.ID(u.idGenerator.NewID()),
		input.Name,
		input.CurrentUserID,
		input.SettlementMode,
		input.Currency,
		u.now(),
	)
	if err != nil {
		return inbound.RegisterHouseholdOutput{}, fmt.Errorf("register household: %w", err)
	}

	if err := u.repo.Save(ctx, h); err != nil {
		return inbound.RegisterHouseholdOutput{}, fmt.Errorf("register household: %w", err)
	}

	householdID := string(h.ID())
	now := u.now()

	// Seed default categories for the new household.
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

	slog.InfoContext(ctx, "register household", "household_id", householdID, "default_categories_seeded", len(defaultCategoryDefs))
	return inbound.RegisterHouseholdOutput{HouseholdID: householdID}, nil
}
