package householdapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ListHouseholdsUseCase lists households with pagination.
type ListHouseholdsUseCase struct {
	repo outbound.HouseholdRepository
}

// NewListHouseholdsUseCase constructs ListHouseholdsUseCase.
func NewListHouseholdsUseCase(repo outbound.HouseholdRepository) ListHouseholdsUseCase {
	return ListHouseholdsUseCase{repo: repo}
}

// Execute retrieves households with bounded pagination.
func (u ListHouseholdsUseCase) Execute(ctx context.Context, query inbound.ListHouseholdsQuery) ([]household.Household, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	households, err := u.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list households: %w", err)
	}

	return households, nil
}
