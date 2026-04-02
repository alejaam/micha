package cardapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/card"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ListCardsUseCase lists all cards for a household.
type ListCardsUseCase struct {
	repo outbound.CardRepository
}

// NewListCardsUseCase constructs ListCardsUseCase.
func NewListCardsUseCase(repo outbound.CardRepository) ListCardsUseCase {
	return ListCardsUseCase{repo: repo}
}

// Execute retrieves all cards for the specified household.
func (u ListCardsUseCase) Execute(ctx context.Context, query inbound.ListCardsQuery) ([]card.Card, error) {
	cards, err := u.repo.ListByHousehold(ctx, query.HouseholdID)
	if err != nil {
		return nil, fmt.Errorf("list cards: %w", err)
	}
	return cards, nil
}

var _ inbound.ListCardsUseCase = ListCardsUseCase{}
