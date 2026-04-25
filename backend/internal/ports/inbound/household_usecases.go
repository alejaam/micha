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
	CurrentUserID  string
}

// RegisterHouseholdOutput contains created household identifiers.
type RegisterHouseholdOutput struct {
	HouseholdID string
}

// ListHouseholdsQuery holds pagination for listing households.
type ListHouseholdsQuery struct {
	// UserID scopes the query to households the user belongs to.
	UserID string
	Limit  int
	Offset int
}

// UpdateHouseholdInput contains mutable fields for updating a household.
type UpdateHouseholdInput struct {
	HouseholdID    string
	Name           string
	SettlementMode household.SettlementMode
	Currency       string
}

type RegisterHouseholdUseCase interface {
	Execute(ctx context.Context, input RegisterHouseholdInput) (RegisterHouseholdOutput, error)
}

type ListHouseholdsUseCase interface {
	Execute(ctx context.Context, query ListHouseholdsQuery) ([]household.Household, error)
}

type GetHouseholdUseCase interface {
	Execute(ctx context.Context, householdID string) (household.Household, error)
}

type UpdateHouseholdUseCase interface {
	Execute(ctx context.Context, input UpdateHouseholdInput) error
}
