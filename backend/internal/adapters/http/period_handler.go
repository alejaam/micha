package httpadapter

import (
	"errors"
	"net/http"

	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type PeriodHandlerDeps struct {
	SimulateClose    inbound.SimulateClosePeriodUseCase
	InitializePeriod inbound.InitializePeriodUseCase
	PeriodRepo       outbound.PeriodRepository
}

type PeriodHandler struct {
	simulateClose    inbound.SimulateClosePeriodUseCase
	initializePeriod inbound.InitializePeriodUseCase
	periodRepo       outbound.PeriodRepository
}

func newPeriodHandler(deps PeriodHandlerDeps) *PeriodHandler {
	return &PeriodHandler{
		simulateClose:    deps.SimulateClose,
		initializePeriod: deps.InitializePeriod,
		periodRepo:       deps.PeriodRepo,
	}
}

func (h *PeriodHandler) handleInitialize(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	userID, _ := UserIDFromContext(r.Context())

	output, err := h.initializePeriod.Execute(r.Context(), inbound.InitializePeriodInput{
		HouseholdID:   householdID,
		CurrentUserID: userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, shared.ErrForbidden):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, output)
}

func (h *PeriodHandler) handleSimulateClose(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	periodID := r.PathValue("period_id")
	userID, _ := UserIDFromContext(r.Context())

	output, err := h.simulateClose.Execute(r.Context(), inbound.SimulateClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: userID,
	})
	if err != nil {
		switch {
		case errors.Is(err, shared.ErrForbidden):
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, shared.ErrFuturePeriod):
			writeError(w, http.StatusBadRequest, "FUTURE_PERIOD", "next period would start in the future")
		case errors.Is(err, period.ErrPeriodTooShort):
			writeError(w, http.StatusBadRequest, "PERIOD_TOO_SHORT", "period must be open for at least 7 days before closing")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"next_period_start":   output.NextPeriodStart,
			"next_period_end":     output.NextPeriodEnd,
			"settlement_preview":  output.SettlementPreview,
			"fixed_expense_count": output.FixedExpenseCount,
			"installment_count":   output.InstallmentCount,
		},
	})
}

func (h *PeriodHandler) handleGetCurrent(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")

	p, err := h.periodRepo.GetLatestByHousehold(r.Context(), householdID)
	if err != nil {
		if err.Error() == "not found" || err.Error() == "no rows in result set" {
			writeJSON(w, http.StatusOK, map[string]any{"data": nil})
			return
		}
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	attrs := p.Attributes()

	// Natural Closing: If period is OPEN but EndDate has passed, just report it as is.
	// The review/approval workflow has been removed — the period stays open until
	// a future close mechanism is implemented.

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":           string(attrs.ID),
			"household_id": attrs.HouseholdID,
			"start_date":   attrs.StartDate,
			"end_date":     attrs.EndDate,
			"status":       string(attrs.Status),
			"created_at":   attrs.CreatedAt,
			"updated_at":   attrs.UpdatedAt,
		},
	})
}

func (h *PeriodHandler) handleListHistory(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	periods, err := h.periodRepo.ListByHousehold(r.Context(), householdID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	items := make([]map[string]any, 0, len(periods))
	for _, p := range periods {
		attrs := p.Attributes()
		items = append(items, map[string]any{
			"id":           string(attrs.ID),
			"household_id": attrs.HouseholdID,
			"start_date":   attrs.StartDate,
			"end_date":     attrs.EndDate,
			"status":       string(attrs.Status),
			"created_at":   attrs.CreatedAt,
			"updated_at":   attrs.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}
