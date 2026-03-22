package categoryapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// CreateCategoryUseCase creates a custom expense category for a household.
type CreateCategoryUseCase struct {
	repo  outbound.CategoryRepository
	idGen appshared.IDGenerator
	now   func() time.Time
}

// NewCreateCategoryUseCase constructs a CreateCategoryUseCase.
func NewCreateCategoryUseCase(repo outbound.CategoryRepository, idGen appshared.IDGenerator) CreateCategoryUseCase {
	return CreateCategoryUseCase{repo: repo, idGen: idGen, now: time.Now}
}

func (u CreateCategoryUseCase) Execute(ctx context.Context, input inbound.CreateCategoryInput) (inbound.CreateCategoryOutput, error) {
	// Reject if slug already exists in this household.
	if _, err := u.repo.FindBySlug(ctx, input.HouseholdID, input.Slug); err == nil {
		return inbound.CreateCategoryOutput{}, fmt.Errorf("create category: %w", shared.ErrAlreadyExists)
	}

	c, err := category.New(
		u.idGen.NewID(),
		input.HouseholdID,
		input.Name,
		input.Slug,
		u.now(),
	)
	if err != nil {
		return inbound.CreateCategoryOutput{}, fmt.Errorf("create category: %w", err)
	}

	if err := u.repo.Save(ctx, c); err != nil {
		return inbound.CreateCategoryOutput{}, fmt.Errorf("create category: %w", err)
	}

	slog.InfoContext(ctx, "create category", "category_id", string(c.ID()), "household_id", input.HouseholdID)
	return inbound.CreateCategoryOutput{CategoryID: string(c.ID())}, nil
}

var _ inbound.CreateCategoryUseCase = CreateCategoryUseCase{}
