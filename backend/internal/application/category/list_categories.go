package categoryapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/category"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ListCategoriesUseCase returns all categories (default + custom) for a household.
type ListCategoriesUseCase struct {
	repo outbound.CategoryRepository
}

// NewListCategoriesUseCase constructs a ListCategoriesUseCase.
func NewListCategoriesUseCase(repo outbound.CategoryRepository) ListCategoriesUseCase {
	return ListCategoriesUseCase{repo: repo}
}

func (u ListCategoriesUseCase) Execute(ctx context.Context, query inbound.ListCategoriesQuery) ([]category.Category, error) {
	cats, err := u.repo.ListByHousehold(ctx, query.HouseholdID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}

var _ inbound.ListCategoriesUseCase = ListCategoriesUseCase{}
