package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/subscriptioncatalog"
	"micha/backend/internal/ports/inbound"
)

// SubscriptionCatalogHandlerDeps groups all use case dependencies for the subscription catalog.
type SubscriptionCatalogHandlerDeps struct {
	ListServices       inbound.ListCatalogServicesUseCase
	LinkService        inbound.LinkServiceUseCase
	UnlinkService      inbound.UnlinkServiceUseCase
	GetLinksByExpense  inbound.ListLinksByExpenseUseCase
	GetSubscriptionKPI inbound.GetSubscriptionKPIUseCase
}

type subscriptionCatalogHandler struct {
	deps SubscriptionCatalogHandlerDeps
}

func newSubscriptionCatalogHandler(deps SubscriptionCatalogHandlerDeps) subscriptionCatalogHandler {
	return subscriptionCatalogHandler{deps: deps}
}

// handleListServices handles GET /v1/subscription-services.
func (h subscriptionCatalogHandler) handleListServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.deps.ListServices.Execute(r.Context())
	if err != nil {
		slog.Error("subscription catalog handler: list services", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list subscription services")
		return
	}

	items := make([]map[string]any, 0, len(services))
	for _, s := range services {
		items = append(items, subscriptionServiceJSON(s))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// handleLinkService handles POST /v1/recurring-expenses/{recurring_expense_id}/catalog-links.
func (h subscriptionCatalogHandler) handleLinkService(w http.ResponseWriter, r *http.Request) {
	recurringExpenseID := r.PathValue("recurring_expense_id")
	if _, err := uuid.Parse(recurringExpenseID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "recurring_expense_id must be a valid UUID")
		return
	}

	var body struct {
		CatalogServiceID string `json:"catalog_service_id"`
		CustomPriceCents *int64 `json:"custom_price_cents"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}
	if _, err := uuid.Parse(body.CatalogServiceID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CATALOG_SERVICE_ID", "catalog_service_id must be a valid UUID")
		return
	}

	link, err := h.deps.LinkService.Execute(r.Context(), inbound.LinkServiceInput{
		RecurringExpenseID: recurringExpenseID,
		CatalogServiceID:   body.CatalogServiceID,
		CustomPriceCents:   body.CustomPriceCents,
	})
	if err != nil {
		writeCatalogLinkError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": catalogLinkJSON(link)})
}

// handleUnlinkService handles DELETE /v1/recurring-expenses/{recurring_expense_id}/catalog-links/{catalog_service_id}.
func (h subscriptionCatalogHandler) handleUnlinkService(w http.ResponseWriter, r *http.Request) {
	recurringExpenseID := r.PathValue("recurring_expense_id")
	catalogServiceID := r.PathValue("catalog_service_id")

	if _, err := uuid.Parse(recurringExpenseID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "recurring_expense_id must be a valid UUID")
		return
	}
	if _, err := uuid.Parse(catalogServiceID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CATALOG_SERVICE_ID", "catalog_service_id must be a valid UUID")
		return
	}

	if err := h.deps.UnlinkService.Execute(r.Context(), recurringExpenseID, catalogServiceID); err != nil {
		writeCatalogLinkError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListExpenseLinks handles GET /v1/recurring-expenses/{recurring_expense_id}/catalog-links.
func (h subscriptionCatalogHandler) handleListExpenseLinks(w http.ResponseWriter, r *http.Request) {
	recurringExpenseID := r.PathValue("recurring_expense_id")
	if _, err := uuid.Parse(recurringExpenseID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "recurring_expense_id must be a valid UUID")
		return
	}

	links, err := h.deps.GetLinksByExpense.Execute(r.Context(), recurringExpenseID)
	if err != nil {
		slog.Error("subscription catalog handler: list expense links", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list catalog links")
		return
	}

	items := make([]map[string]any, 0, len(links))
	for _, l := range links {
		items = append(items, catalogLinkJSON(l))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// handleGetKPI handles GET /v1/households/{household_id}/subscription-kpi.
func (h subscriptionCatalogHandler) handleGetKPI(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	if householdID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "household_id is required")
		return
	}

	kpi, err := h.deps.GetSubscriptionKPI.Execute(r.Context(), householdID)
	if err != nil {
		slog.Error("subscription catalog handler: get KPI", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate subscription KPI")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": kpi})
}

// --- JSON helpers ---

func subscriptionServiceJSON(s subscriptioncatalog.SubscriptionService) map[string]any {
	attrs := s.Attributes()
	m := map[string]any{
		"id":                      string(attrs.ID),
		"name":                    attrs.Name,
		"slug":                    attrs.Slug,
		"region":                  attrs.Region,
		"currency":                attrs.Currency,
		"standalone_price_cents":  attrs.StandalonePriceCents,
		"icon_url":                attrs.IconURL,
		"is_bundle":               attrs.IsBundle,
		"created_at":              attrs.CreatedAt,
	}
	return m
}

func catalogLinkJSON(l subscriptioncatalog.CatalogLink) map[string]any {
	attrs := l.Attributes()
	m := map[string]any{
		"recurring_expense_id": attrs.RecurringExpenseID,
		"catalog_service_id":   attrs.CatalogServiceID,
		"custom_price_cents":   attrs.CustomPriceCents,
		"created_at":           attrs.CreatedAt.Format(time.RFC3339),
	}
	return m
}

func writeCatalogLinkError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "the requested resource was not found")
	case errors.Is(err, subscriptioncatalog.ErrLinkAlreadyExists):
		writeError(w, http.StatusConflict, "LINK_ALREADY_EXISTS", "service is already linked to this recurring expense")
	case errors.Is(err, subscriptioncatalog.ErrInvalidCustomPrice):
		writeError(w, http.StatusBadRequest, "INVALID_CUSTOM_PRICE", "custom_price_cents must be greater than zero when provided")
	default:
		slog.Error("subscription catalog handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}


