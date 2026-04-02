package cardapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// DeleteCardUseCase soft-deletes a card.
type DeleteCardUseCase struct {
	repo outbound.CardRepository
}

// NewDeleteCardUseCase constructs DeleteCardUseCase.
func NewDeleteCardUseCase(repo outbound.CardRepository) DeleteCardUseCase {
	return DeleteCardUseCase{repo: repo}
}

// Execute soft-deletes the specified card.
// Returns shared.ErrNotFound if the card does not exist or is already deleted.
func (u DeleteCardUseCase) Execute(ctx context.Context, input inbound.DeleteCardInput) error {
	// Verify the card exists and belongs to the household
	c, err := u.repo.FindByID(ctx, input.CardID)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	if c.HouseholdID() != input.HouseholdID {
		return shared.ErrNotFound
	}

	if err := u.repo.Delete(ctx, input.CardID); err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	return nil
}

var _ inbound.DeleteCardUseCase = DeleteCardUseCase{}
