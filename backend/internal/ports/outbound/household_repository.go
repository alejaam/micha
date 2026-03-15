package outbound

import (
	"context"

	"micha/backend/internal/domain/household"
)

// HouseholdRepository defines persistence operations required by household use cases.
type HouseholdRepository interface {
	Save(ctx context.Context, h household.Household) error
	FindByID(ctx context.Context, id string) (household.Household, error)
	// List returns all households (admin/internal use only). Use ListByUserID for user-scoped queries.
	List(ctx context.Context, limit, offset int) ([]household.Household, error)
	// ListByUserID returns only the households the given user belongs to (via members table).
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]household.Household, error)
	Update(ctx context.Context, h household.Household) error
}
