package httpadapter

import (
	"net/http"

	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type PeriodHandlerDeps struct {
	TransitionToReview inbound.TransitionToReviewUseCase
	ApprovePeriod       inbound.ApprovePeriodUseCase
	ClosePeriod         inbound.ClosePeriodUseCase
	PeriodRepo          outbound.PeriodRepository
}

type PeriodHandler struct {
	transitionToReview inbound.TransitionToReviewUseCase
	approvePeriod       inbound.ApprovePeriodUseCase
	closePeriod         inbound.ClosePeriodUseCase
	periodRepo          outbound.PeriodRepository
}

func newPeriodHandler(deps PeriodHandlerDeps) *PeriodHandler {
	return &PeriodHandler{
		transitionToReview: deps.TransitionToReview,
		approvePeriod:       deps.ApprovePeriod,
		closePeriod:         deps.ClosePeriod,
		periodRepo:          deps.PeriodRepo,
	}
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

	p, err := h.periodRepo.GetCurrentOpen(r.Context(), householdID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "no open period found")
		return
	}

	writeJSON(w, http.StatusOK, p.Attributes())
}
