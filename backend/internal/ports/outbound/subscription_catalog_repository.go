package outbound

import (
	"context"

	"micha/backend/internal/domain/subscriptioncatalog"
)

// SubscriptionCatalogRepository defines the persistence contract for the subscription catalog.
type SubscriptionCatalogRepository interface {
	// ListServices returns all catalog services.
	ListServices(ctx context.Context) ([]subscriptioncatalog.SubscriptionService, error)

	// FindServiceByID retrieves a single catalog service.
	FindServiceByID(ctx context.Context, id string) (subscriptioncatalog.SubscriptionService, error)

	// CreateLink links a recurring expense to a catalog service.
	CreateLink(ctx context.Context, link subscriptioncatalog.CatalogLink) error

	// DeleteLink removes a link between a recurring expense and a catalog service.
	DeleteLink(ctx context.Context, recurringExpenseID, catalogServiceID string) error

	// ListLinksByExpense returns all catalog links for a given recurring expense.
	ListLinksByExpense(ctx context.Context, recurringExpenseID string) ([]subscriptioncatalog.CatalogLink, error)

	// ListLinksByHousehold returns all catalog links for a household's recurring expenses,
	// joined with service prices for KPI calculation.
	ListLinksByHousehold(ctx context.Context, householdID string) ([]RecurringExpenseLink, error)
}

// RecurringExpenseLink is a joined result from recurring_expense + catalog_link + subscription_service.
type RecurringExpenseLink struct {
	RecurringExpenseID string
	ServiceID          string
	ServiceName        string
	StandalonePriceCents int64
	CustomPriceCents   *int64
	ExpenseAmountCents int64
}
