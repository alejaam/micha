package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// MemberFinanceHandlerDeps groups use cases for member finance reporting.
type MemberFinanceHandlerDeps struct {
	CalculateRemainingSalary inbound.CalculateRemainingSalaryUseCase
}

type memberFinanceHandler struct {
	deps MemberFinanceHandlerDeps
}

func newMemberFinanceHandler(deps MemberFinanceHandlerDeps) memberFinanceHandler {
	return memberFinanceHandler{deps: deps}
}

// handleGetRemainingSalary handles
// GET /v1/households/{household_id}/members/{member_id}/remaining-salary?year=&month=.
func (h memberFinanceHandler) handleGetRemainingSalary(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
		return
	}

	memberID := r.PathValue("member_id")
	if _, err := uuid.Parse(memberID); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_MEMBER_ID", "member_id must be a valid UUID")
		return
	}

	now := time.Now().UTC()
	year := queryInt(r, "year", now.Year())
	month := queryInt(r, "month", int(now.Month()))
	if year < 2000 || year > 2200 {
		writeError(w, http.StatusBadRequest, "INVALID_YEAR", "year must be between 2000 and 2200")
		return
	}
	if month < 1 || month > 12 {
		writeError(w, http.StatusBadRequest, "INVALID_MONTH", "month must be between 1 and 12")
		return
	}

	out, err := h.deps.CalculateRemainingSalary.Execute(r.Context(), inbound.CalculateRemainingSalaryInput{
		HouseholdID: householdID,
		MemberID:    memberID,
		Year:        year,
		Month:       month,
	})
	if err != nil {
		writeErrorFromMemberFinanceDomain(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"household_id":            out.HouseholdID,
			"member_id":               out.MemberID,
			"year":                    out.Year,
			"month":                   out.Month,
			"monthly_salary_cents":    out.MonthlySalaryCents,
			"personal_expenses_cents": out.PersonalExpensesCents,
			"allocated_debt_cents":    out.AllocatedDebtCents,
			"remaining_salary_cents":  out.RemainingSalaryCents,
		},
	})
}

func writeErrorFromMemberFinanceDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")
	case errors.Is(err, shared.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "you are not allowed to access this member")
	case errors.Is(err, shared.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
	default:
		slog.Error("member finance handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
