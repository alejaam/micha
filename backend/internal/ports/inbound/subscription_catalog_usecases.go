package inbound

import (
	"context"

	"micha/backend/internal/domain/subscriptioncatalog"
)

// ListCatalogServicesUseCase returns all services in the subscription catalog.
type ListCatalogServicesUseCase interface {
	Execute(ctx context.Context) ([]subscriptioncatalog.SubscriptionService, error)
}

// LinkServiceInput contains the data needed to link a service to a recurring expense.
type LinkServiceInput struct {
	RecurringExpenseID string
	CatalogServiceID   string
	CustomPriceCents   *int64
}

// LinkServiceUseCase links a catalog service to a recurring expense.
type LinkServiceUseCase interface {
	Execute(ctx context.Context, input LinkServiceInput) (subscriptioncatalog.CatalogLink, error)
}

// ListLinksByExpenseUseCase returns all catalog links for a recurring expense.
type ListLinksByExpenseUseCase interface {
	Execute(ctx context.Context, recurringExpenseID string) ([]subscriptioncatalog.CatalogLink, error)
}

// UnlinkServiceUseCase removes a catalog service link from a recurring expense.
type UnlinkServiceUseCase interface {
	Execute(ctx context.Context, recurringExpenseID, catalogServiceID string) error
}

// SubscriptionKPIOutput contains the KPI calculation result.
type SubscriptionKPIOutput struct {
	TotalSpentCents      int64              `json:"total_spent_cents"`
	StandaloneTotalCents int64              `json:"standalone_total_cents"`
	NetSavingsCents      int64              `json:"net_savings_cents"`
	OverlapWarnings      []OverlapWarning   `json:"overlap_warnings"`
}

// OverlapWarning indicates a catalog service linked to multiple recurring expenses.
type OverlapWarning struct {
	CatalogServiceID   string   `json:"catalog_service_id"`
	ServiceName        string   `json:"service_name"`
	RecurringExpenseIDs []string `json:"recurring_expense_ids"`
}

// GetSubscriptionKPIUseCase calculates KPI for subscription-linked expenses.
type GetSubscriptionKPIUseCase interface {
	Execute(ctx context.Context, householdID string) (SubscriptionKPIOutput, error)
}
