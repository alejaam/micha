package outbound

import (
	"context"
	"micha/backend/internal/domain/period"
)

// PeriodRepository defines the contract for persisting and retrieving periods.
type PeriodRepository interface {
	// Create persists a new period.
	Create(ctx context.Context, p period.Period) error

	// Update updates an existing period.
	Update(ctx context.Context, p period.Period) error

	// GetByID retrieves a period by its unique identifier.
	GetByID(ctx context.Context, id period.ID) (period.Period, error)

	// GetCurrentOpen retrieves the currently open period for a household.
	GetCurrentOpen(ctx context.Context, householdID string) (period.Period, error)

	// GetLatestByHousehold retrieves the latest period (regardless of status) for a household.
	GetLatestByHousehold(ctx context.Context, householdID string) (period.Period, error)
}
