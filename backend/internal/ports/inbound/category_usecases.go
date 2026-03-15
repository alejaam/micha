package inbound

import (
	"context"

	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/household"
)

// --- Category use cases ---

// CreateCategoryInput contains the required data to create a custom category.
type CreateCategoryInput struct {
	HouseholdID string
	Name        string
	Slug        string
}

// CreateCategoryOutput contains the created category ID.
type CreateCategoryOutput struct {
	CategoryID string
}

// CreateCategoryUseCase creates a custom category for a household.
type CreateCategoryUseCase interface {
	Execute(ctx context.Context, input CreateCategoryInput) (CreateCategoryOutput, error)
}

// ListCategoriesQuery holds parameters for listing categories.
type ListCategoriesQuery struct {
	HouseholdID string
}

// ListCategoriesUseCase lists all categories (default + custom) for a household.
type ListCategoriesUseCase interface {
	Execute(ctx context.Context, query ListCategoriesQuery) ([]category.Category, error)
}

// DeleteCategoryInput identifies the category to delete.
type DeleteCategoryInput struct {
	HouseholdID string
	CategoryID  string
}

// DeleteCategoryUseCase deletes a custom category.
type DeleteCategoryUseCase interface {
	Execute(ctx context.Context, input DeleteCategoryInput) error
}

// --- Household split config use case ---

// UpdateSplitConfigInput carries the desired per-member split percentages.
type UpdateSplitConfigInput struct {
	HouseholdID string
	Splits      []household.MemberSplit
}

// UpdateSplitConfigUseCase sets or replaces the household's split configuration.
type UpdateSplitConfigUseCase interface {
	Execute(ctx context.Context, input UpdateSplitConfigInput) error
}
