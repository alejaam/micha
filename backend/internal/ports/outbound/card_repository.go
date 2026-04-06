package outbound

import (
	"context"

	"micha/backend/internal/domain/card"
)

// CardRepository defines persistence operations required by card use cases.
type CardRepository interface {
	Save(ctx context.Context, c card.Card) error
	FindByID(ctx context.Context, id string) (card.Card, error)
	ListByHousehold(ctx context.Context, householdID string) ([]card.Card, error)
	// Delete soft-deletes a card by setting deleted_at.
	Delete(ctx context.Context, id string) error
}
