package inbound

import (
	"context"

	"micha/backend/internal/domain/card"
)

// RegisterCardInput contains required data to register a card.
type RegisterCardInput struct {
	HouseholdID   string
	OwnerMemberID string
	BankName      string
	CardName      string
	CutoffDay     int
}

// RegisterCardOutput contains the created card identifier.
type RegisterCardOutput struct {
	CardID string
}

// ListCardsQuery holds listing parameters for cards by household.
type ListCardsQuery struct {
	HouseholdID string
}

// DeleteCardInput contains data required to delete a card.
type DeleteCardInput struct {
	CardID      string
	HouseholdID string
}

type RegisterCardUseCase interface {
	Execute(ctx context.Context, input RegisterCardInput) (RegisterCardOutput, error)
}

type ListCardsUseCase interface {
	Execute(ctx context.Context, query ListCardsQuery) ([]card.Card, error)
}

type DeleteCardUseCase interface {
	Execute(ctx context.Context, input DeleteCardInput) error
}
