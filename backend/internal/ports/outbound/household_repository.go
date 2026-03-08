package outbound

import (
	"context"

	"micha/backend/internal/domain/household"
)

// HouseholdRepository defines persistence operations required by household use cases.
type HouseholdRepository interface {
	Save(ctx context.Context, h household.Household) error
	FindByID(ctx context.Context, id string) (household.Household, error)
	List(ctx context.Context, limit, offset int) ([]household.Household, error)
	Update(ctx context.Context, h household.Household) error
}
