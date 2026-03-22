package outbound

import (
	"context"

	"micha/backend/internal/domain/category"
)

// CategoryRepository defines the persistence contract for expense categories.
type CategoryRepository interface {
	// Save persists a new category.
	Save(ctx context.Context, c category.Category) error
	// FindBySlug retrieves a category by household and slug.
	FindBySlug(ctx context.Context, householdID, slug string) (category.Category, error)
	// ListByHousehold returns all categories for a household (default + custom).
	ListByHousehold(ctx context.Context, householdID string) ([]category.Category, error)
	// Delete removes a custom category by ID. Must not delete default categories.
	Delete(ctx context.Context, id string) error
}
