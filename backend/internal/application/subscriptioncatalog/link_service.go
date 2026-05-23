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

// LinkServiceUseCase implements inbound.LinkServiceUseCase.
type LinkServiceUseCase struct {
	repo outbound.SubscriptionCatalogRepository
}

// NewLinkServiceUseCase constructs a new LinkServiceUseCase.
func NewLinkServiceUseCase(repo outbound.SubscriptionCatalogRepository) LinkServiceUseCase {
	return LinkServiceUseCase{repo: repo}
}

// Execute links a catalog service to a recurring expense.
func (u LinkServiceUseCase) Execute(ctx context.Context, input inbound.LinkServiceInput) (subscriptioncatalog.CatalogLink, error) {
	// Validate the service exists
	_, err := u.repo.FindServiceByID(ctx, input.CatalogServiceID)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return subscriptioncatalog.CatalogLink{}, subscriptioncatalog.ErrCatalogServiceNotFound
		}
		return subscriptioncatalog.CatalogLink{}, fmt.Errorf("link service: %w", err)
	}

	// Check for existing links to prevent duplicates
	existing, err := u.repo.ListLinksByExpense(ctx, input.RecurringExpenseID)
	if err != nil {
		return subscriptioncatalog.CatalogLink{}, fmt.Errorf("link service: %w", err)
	}
	for _, link := range existing {
		if link.CatalogServiceID() == input.CatalogServiceID {
			return subscriptioncatalog.CatalogLink{}, subscriptioncatalog.ErrLinkAlreadyExists
		}
	}

	link, err := subscriptioncatalog.NewCatalogLink(subscriptioncatalog.CatalogLinkAttributes{
		RecurringExpenseID: input.RecurringExpenseID,
		CatalogServiceID:   input.CatalogServiceID,
		CustomPriceCents:   input.CustomPriceCents,
	})
	if err != nil {
		return subscriptioncatalog.CatalogLink{}, err
	}

	if err := u.repo.CreateLink(ctx, link); err != nil {
		return subscriptioncatalog.CatalogLink{}, fmt.Errorf("link service: %w", err)
	}

	return link, nil
}

var _ inbound.LinkServiceUseCase = LinkServiceUseCase{}
