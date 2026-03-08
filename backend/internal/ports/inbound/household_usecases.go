package inbound

import (
	"context"

	"micha/backend/internal/domain/household"
)

// RegisterHouseholdInput contains required data to register a household.
type RegisterHouseholdInput struct {
	Name           string
	SettlementMode household.SettlementMode
	Currency       string
}

// RegisterHouseholdOutput contains created household identifiers.
type RegisterHouseholdOutput struct {
	HouseholdID string
}

// ListHouseholdsQuery holds pagination for listing households.
type ListHouseholdsQuery struct {
	Limit  int
	Offset int
}

type RegisterHouseholdUseCase interface {
	Execute(ctx context.Context, input RegisterHouseholdInput) (RegisterHouseholdOutput, error)
}

type ListHouseholdsUseCase interface {
	Execute(ctx context.Context, query ListHouseholdsQuery) ([]household.Household, error)
}
