package householdapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

var _ inbound.GetHouseholdUseCase = GetHouseholdUseCase{}

// GetHouseholdUseCase retrieves a household by ID.
type GetHouseholdUseCase struct {
	repo outbound.HouseholdRepository
}

// NewGetHouseholdUseCase constructs GetHouseholdUseCase.
func NewGetHouseholdUseCase(repo outbound.HouseholdRepository) GetHouseholdUseCase {
	return GetHouseholdUseCase{repo: repo}
}

// Execute retrieves the household.
func (u GetHouseholdUseCase) Execute(ctx context.Context, householdID string) (household.Household, error) {
	h, err := u.repo.FindByID(ctx, householdID)
	if err != nil {
		return household.Household{}, fmt.Errorf("get household: %w", err)
	}
	return h, nil
}
