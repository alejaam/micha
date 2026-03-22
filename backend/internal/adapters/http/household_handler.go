package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// HouseholdHandlerDeps groups all use case dependencies for the household resource.
type HouseholdHandlerDeps struct {
	Register inbound.RegisterHouseholdUseCase
	List     inbound.ListHouseholdsUseCase
	Get      inbound.GetHouseholdUseCase
	Update   inbound.UpdateHouseholdUseCase
}

// householdHandler handles HTTP requests for the household resource.
type householdHandler struct {
	deps HouseholdHandlerDeps
}

func newHouseholdHandler(deps HouseholdHandlerDeps) householdHandler {
	return householdHandler{deps: deps}
}

// handleCreate handles POST /v1/households.
func (h householdHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name           string `json:"name"`
		SettlementMode string `json:"settlement_mode"`
		Currency       string `json:"currency"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	out, err := h.deps.Register.Execute(r.Context(), inbound.RegisterHouseholdInput{
		Name:           body.Name,
		SettlementMode: household.SettlementMode(body.SettlementMode),
		Currency:       body.Currency,
	})
	if err != nil {
		writeErrorFromHouseholdDomain(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]string{"household_id": out.HouseholdID},
	})
}

// handleList handles GET /v1/households?limit=&offset=.
func (h householdHandler) handleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	households, err := h.deps.List.Execute(r.Context(), inbound.ListHouseholdsQuery{
		UserID: userID,
		Limit:  queryInt(r, "limit", 20),
		Offset: queryInt(r, "offset", 0),
	})
	if err != nil {
		writeErrorFromHouseholdDomain(w, err)
		return
	}

	items := make([]map[string]any, 0, len(households))
	for _, item := range households {
		attrs := item.Attributes()
		items = append(items, map[string]any{
			"id":              string(attrs.ID),
			"name":            attrs.Name,
			"settlement_mode": attrs.SettlementMode,
			"currency":        attrs.Currency,
			"created_at":      attrs.CreatedAt,
			"updated_at":      attrs.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func writeErrorFromHouseholdDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "household not found")
	case errors.Is(err, household.ErrInvalidName):
		writeError(w, http.StatusBadRequest, "INVALID_NAME", "household name is required")
	case errors.Is(err, household.ErrInvalidSettlementMode):
		writeError(w, http.StatusBadRequest, "INVALID_SETTLEMENT_MODE", "settlement_mode must be equal or proportional")
	case errors.Is(err, household.ErrInvalidCurrency):
		writeError(w, http.StatusBadRequest, "INVALID_CURRENCY", "currency must be a 3-letter code")
	default:
		slog.Error("household handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}

// handleGet handles GET /v1/households/{household_id}.
func (h householdHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	if _, err := uuid.Parse(householdID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HOUSEHOLD_ID", "household_id must be a valid UUID")
		return
	}

	hh, err := h.deps.Get.Execute(r.Context(), householdID)
	if err != nil {
		writeErrorFromHouseholdDomain(w, err)
		return
	}

	attrs := hh.Attributes()
	splitConfig := make([]map[string]any, 0, len(attrs.SplitConfig.Splits()))
	for _, s := range attrs.SplitConfig.Splits() {
		splitConfig = append(splitConfig, map[string]any{
			"member_id":  s.MemberID,
			"percentage": s.Percentage,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":              string(attrs.ID),
			"name":            attrs.Name,
			"settlement_mode": attrs.SettlementMode,
			"currency":        attrs.Currency,
			"split_config":    splitConfig,
			"created_at":      attrs.CreatedAt,
			"updated_at":      attrs.UpdatedAt,
		},
	})
}

// handleUpdate handles PUT /v1/households/{household_id}.
func (h householdHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	if _, err := uuid.Parse(householdID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HOUSEHOLD_ID", "household_id must be a valid UUID")
		return
	}

	var body struct {
		Name           string `json:"name"`
		SettlementMode string `json:"settlement_mode"`
		Currency       string `json:"currency"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	err := h.deps.Update.Execute(r.Context(), inbound.UpdateHouseholdInput{
		HouseholdID:    householdID,
		Name:           body.Name,
		SettlementMode: household.SettlementMode(body.SettlementMode),
		Currency:       body.Currency,
	})
	if err != nil {
		writeErrorFromHouseholdDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
