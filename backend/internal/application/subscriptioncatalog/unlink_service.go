package subscriptioncatalogapp

import (
	"context"
	"errors"
	"fmt"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/subscriptioncatalog"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// UnlinkServiceUseCase implements inbound.UnlinkServiceUseCase.
type UnlinkServiceUseCase struct {
	repo outbound.SubscriptionCatalogRepository
}

// NewUnlinkServiceUseCase constructs a new UnlinkServiceUseCase.
func NewUnlinkServiceUseCase(repo outbound.SubscriptionCatalogRepository) UnlinkServiceUseCase {
	return UnlinkServiceUseCase{repo: repo}
}

// Execute removes a catalog service link from a recurring expense.
func (u UnlinkServiceUseCase) Execute(ctx context.Context, recurringExpenseID, catalogServiceID string) error {
	err := u.repo.DeleteLink(ctx, recurringExpenseID, catalogServiceID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return subscriptioncatalog.ErrLinkNotFound
		}
		return fmt.Errorf("unlink service: %w", err)
	}
	return nil
}

var _ inbound.UnlinkServiceUseCase = UnlinkServiceUseCase{}
