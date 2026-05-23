package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/subscriptioncatalog"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

// SubscriptionCatalogRepository fulfills outbound.SubscriptionCatalogRepository using PostgreSQL.
type SubscriptionCatalogRepository struct {
	db *pgxpool.Pool
}

// NewSubscriptionCatalogRepository constructs a new SubscriptionCatalogRepository.
func NewSubscriptionCatalogRepository(db *pgxpool.Pool) SubscriptionCatalogRepository {
	return SubscriptionCatalogRepository{db: db}
}

// ListServices returns all subscription services ordered by name.
func (r SubscriptionCatalogRepository) ListServices(ctx context.Context) ([]subscriptioncatalog.SubscriptionService, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, slug, region, currency, standalone_price_cents,
		        COALESCE(icon_url, ''), is_bundle, created_at, updated_at
		 FROM subscription_services
		 ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("subscription catalog repository listServices: %w", err)
	}
	defer rows.Close()

	var services []subscriptioncatalog.SubscriptionService
	for rows.Next() {
		s, err := scanSubscriptionService(rows)
		if err != nil {
			return nil, fmt.Errorf("subscription catalog repository listServices: scan: %w", err)
		}
		services = append(services, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("subscription catalog repository listServices: rows: %w", err)
	}
	return services, nil
}

// FindServiceByID retrieves a single subscription service by ID.
func (r SubscriptionCatalogRepository) FindServiceByID(ctx context.Context, id string) (subscriptioncatalog.SubscriptionService, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, slug, region, currency, standalone_price_cents,
		        COALESCE(icon_url, ''), is_bundle, created_at, updated_at
		 FROM subscription_services
		 WHERE id = $1`,
		id,
	)
	s, err := scanSubscriptionService(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return subscriptioncatalog.SubscriptionService{}, shared.ErrNotFound
	}
	if err != nil {
		return subscriptioncatalog.SubscriptionService{}, fmt.Errorf("subscription catalog repository findServiceByID: %w", err)
	}
	return s, nil
}

// CreateLink creates a link between a recurring expense and a catalog service.
func (r SubscriptionCatalogRepository) CreateLink(ctx context.Context, link subscriptioncatalog.CatalogLink) error {
	attrs := link.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO recurring_expense_catalog_links (recurring_expense_id, catalog_service_id, custom_price_cents, created_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (recurring_expense_id, catalog_service_id) DO NOTHING`,
		attrs.RecurringExpenseID, attrs.CatalogServiceID, attrs.CustomPriceCents, attrs.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("subscription catalog repository createLink: %w", err)
	}
	return nil
}

// DeleteLink removes a link between a recurring expense and a catalog service.
func (r SubscriptionCatalogRepository) DeleteLink(ctx context.Context, recurringExpenseID, catalogServiceID string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM recurring_expense_catalog_links
		 WHERE recurring_expense_id = $1 AND catalog_service_id = $2`,
		recurringExpenseID, catalogServiceID,
	)
	if err != nil {
		return fmt.Errorf("subscription catalog repository deleteLink: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// ListLinksByExpense returns all catalog links for a given recurring expense.
func (r SubscriptionCatalogRepository) ListLinksByExpense(ctx context.Context, recurringExpenseID string) ([]subscriptioncatalog.CatalogLink, error) {
	rows, err := r.db.Query(ctx,
		`SELECT recurring_expense_id, catalog_service_id, custom_price_cents, created_at
		 FROM recurring_expense_catalog_links
		 WHERE recurring_expense_id = $1`,
		recurringExpenseID,
	)
	if err != nil {
		return nil, fmt.Errorf("subscription catalog repository listLinksByExpense: %w", err)
	}
	defer rows.Close()

	var links []subscriptioncatalog.CatalogLink
	for rows.Next() {
		l, err := scanCatalogLink(rows)
		if err != nil {
			return nil, fmt.Errorf("subscription catalog repository listLinksByExpense: scan: %w", err)
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("subscription catalog repository listLinksByExpense: rows: %w", err)
	}
	return links, nil
}

// ListLinksByHousehold returns all catalog links joined with service and expense data for KPI calculation.
func (r SubscriptionCatalogRepository) ListLinksByHousehold(ctx context.Context, householdID string) ([]outbound.RecurringExpenseLink, error) {
	rows, err := r.db.Query(ctx,
		`SELECT cl.recurring_expense_id,
		        cl.catalog_service_id,
		        ss.name,
		        ss.standalone_price_cents,
		        cl.custom_price_cents,
		        re.amount_cents
		 FROM recurring_expense_catalog_links cl
		 JOIN subscription_services ss ON ss.id = cl.catalog_service_id
		 JOIN recurring_expenses re ON re.id = cl.recurring_expense_id
		 WHERE re.household_id = $1
		   AND re.deleted_at IS NULL`,
		householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("subscription catalog repository listLinksByHousehold: %w", err)
	}
	defer rows.Close()

	var links []outbound.RecurringExpenseLink
	for rows.Next() {
		var (
			recurringExpenseID string
			serviceID          string
			serviceName        string
			standalonePrice    int64
			customPrice        *int64
			expenseAmount      int64
		)
		if err := rows.Scan(&recurringExpenseID, &serviceID, &serviceName, &standalonePrice, &customPrice, &expenseAmount); err != nil {
			return nil, fmt.Errorf("subscription catalog repository listLinksByHousehold: scan: %w", err)
		}
		links = append(links, outbound.RecurringExpenseLink{
			RecurringExpenseID:   recurringExpenseID,
			ServiceID:            serviceID,
			ServiceName:          serviceName,
			StandalonePriceCents: standalonePrice,
			CustomPriceCents:     customPrice,
			ExpenseAmountCents:   expenseAmount,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("subscription catalog repository listLinksByHousehold: rows: %w", err)
	}
	return links, nil
}

// Ensure interface compliance.
var _ outbound.SubscriptionCatalogRepository = SubscriptionCatalogRepository{}

// --- Scanners ---

type subscriptionServiceRow interface {
	Scan(dest ...any) error
}

func scanSubscriptionService(r subscriptionServiceRow) (subscriptioncatalog.SubscriptionService, error) {
	var (
		id                   string
		name                 string
		slug                 string
		region               string
		currency             string
		standalonePriceCents int64
		iconURL              string
		isBundle             bool
		createdAt            time.Time
		updatedAt            time.Time
	)
	if err := r.Scan(&id, &name, &slug, &region, &currency, &standalonePriceCents, &iconURL, &isBundle, &createdAt, &updatedAt); err != nil {
		return subscriptioncatalog.SubscriptionService{}, err
	}
	return subscriptioncatalog.NewSubscriptionServiceFromAttributes(subscriptioncatalog.SubscriptionServiceAttributes{
		ID:                   subscriptioncatalog.ID(id),
		Name:                 name,
		Slug:                 slug,
		Region:               region,
		Currency:             currency,
		StandalonePriceCents: standalonePriceCents,
		IconURL:              iconURL,
		IsBundle:             isBundle,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	})
}

func scanCatalogLink(r subscriptionServiceRow) (subscriptioncatalog.CatalogLink, error) {
	var (
		recurringExpenseID string
		catalogServiceID   string
		customPriceCents   *int64
		createdAt          time.Time
	)
	if err := r.Scan(&recurringExpenseID, &catalogServiceID, &customPriceCents, &createdAt); err != nil {
		return subscriptioncatalog.CatalogLink{}, err
	}
	return subscriptioncatalog.NewCatalogLink(subscriptioncatalog.CatalogLinkAttributes{
		RecurringExpenseID: recurringExpenseID,
		CatalogServiceID:   catalogServiceID,
		CustomPriceCents:   customPriceCents,
		CreatedAt:          createdAt,
	})
}
