// Package subscriptioncatalogapp provides use cases for the subscription catalog.
package subscriptioncatalogapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/subscriptioncatalog"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ListCatalogServicesUseCase implements inbound.ListCatalogServicesUseCase.
type ListCatalogServicesUseCase struct {
	repo outbound.SubscriptionCatalogRepository
}

// NewListCatalogServicesUseCase constructs a new ListCatalogServicesUseCase.
func NewListCatalogServicesUseCase(repo outbound.SubscriptionCatalogRepository) ListCatalogServicesUseCase {
	return ListCatalogServicesUseCase{repo: repo}
}

// Execute returns all subscription services from the catalog.
func (u ListCatalogServicesUseCase) Execute(ctx context.Context) ([]subscriptioncatalog.SubscriptionService, error) {
	services, err := u.repo.ListServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("list catalog services: %w", err)
	}
	return services, nil
}

// Ensure interface compliance.
var _ inbound.ListCatalogServicesUseCase = ListCatalogServicesUseCase{}
