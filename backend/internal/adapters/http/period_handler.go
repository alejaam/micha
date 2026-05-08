package httpadapter

import (
	"net/http"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type PeriodHandlerDeps struct {
	TransitionToReview inbound.TransitionToReviewUseCase
	ApprovePeriod      inbound.ApprovePeriodUseCase
	ClosePeriod        inbound.ClosePeriodUseCase
	InitializePeriod   inbound.InitializePeriodUseCase
	PeriodRepo         outbound.PeriodRepository
}

type PeriodHandler struct {
	transitionToReview inbound.TransitionToReviewUseCase
	approvePeriod      inbound.ApprovePeriodUseCase
	closePeriod        inbound.ClosePeriodUseCase
	initializePeriod   inbound.InitializePeriodUseCase
	periodRepo         outbound.PeriodRepository
}

func newPeriodHandler(deps PeriodHandlerDeps) *PeriodHandler {
	return &PeriodHandler{
		transitionToReview: deps.TransitionToReview,
		approvePeriod:      deps.ApprovePeriod,
		closePeriod:        deps.ClosePeriod,
		initializePeriod:   deps.InitializePeriod,
		periodRepo:         deps.PeriodRepo,
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
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, output)
}

func (h *PeriodHandler) handleTransitionToReview(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	periodID := r.PathValue("period_id")
	userID, _ := UserIDFromContext(r.Context())

	output, err := h.transitionToReview.Execute(r.Context(), inbound.TransitionToReviewInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *PeriodHandler) handleApprove(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	periodID := r.PathValue("period_id")
	userID, _ := UserIDFromContext(r.Context())

	var input struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}
	if err := decodeJSON(r, w, &input); err != nil {
		return
	}

	output, err := h.approvePeriod.Execute(r.Context(), inbound.ApprovePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: userID,
		Status:        input.Status,
		Comment:       input.Comment,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *PeriodHandler) handleClose(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	periodID := r.PathValue("period_id")
	userID, _ := UserIDFromContext(r.Context())

	var input struct {
		Force bool `json:"force"`
	}
	if err := decodeJSON(r, w, &input); err != nil {
		// If force is not provided, default to false.
		input.Force = false
	}

	output, err := h.closePeriod.Execute(r.Context(), inbound.ClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: userID,
		Force:         input.Force,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *PeriodHandler) handleGetCurrent(w http.ResponseWriter, r *http.Request) {
	householdID := r.PathValue("household_id")
	userID, _ := UserIDFromContext(r.Context())

	p, err := h.periodRepo.GetLatestByHousehold(r.Context(), householdID)
	if err != nil {
		// If it's a 'not found' error, return 200 with null data
		if err.Error() == "not found" || err.Error() == "no rows in result set" {
			writeJSON(w, http.StatusOK, map[string]any{"data": nil})
			return
		}
		// If it's a real DB error (e.g. missing table), return 500 so we can diagnose
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	attrs := p.Attributes()

	// Natural Closing: If period is OPEN but EndDate has passed, transition to REVIEW automatically.
	if attrs.Status == period.StatusOpen && appshared.Now().After(attrs.EndDate) {
		_, err := h.transitionToReview.Execute(r.Context(), inbound.TransitionToReviewInput{
			HouseholdID:   householdID,
			PeriodID:      string(attrs.ID),
			CurrentUserID: userID,
		})
		if err == nil {
			// Reload period to return the updated status
			if updated, reloadErr := h.periodRepo.GetByID(r.Context(), attrs.ID); reloadErr == nil {
				attrs = updated.Attributes()
			}
		}
	}

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
