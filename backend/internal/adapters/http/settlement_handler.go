package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"micha/backend/internal/domain/settlement"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// SettlementHandlerDeps groups all use case dependencies for settlement reports.
type SettlementHandlerDeps struct {
	Calculate inbound.CalculateSettlementUseCase
}

// settlementHandler handles HTTP requests for settlement reports.
type settlementHandler struct {
	deps SettlementHandlerDeps
}

func newSettlementHandler(deps SettlementHandlerDeps) settlementHandler {
	return settlementHandler{deps: deps}
}

// handleGetMonthly handles GET /v1/households/{household_id}/settlement?year=&month=.
func (h settlementHandler) handleGetMonthly(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
		return
	}

	now := time.Now().UTC()
	year := queryInt(r, "year", now.Year())
	month := queryInt(r, "month", int(now.Month()))

	out, err := h.deps.Calculate.Execute(r.Context(), inbound.CalculateSettlementInput{
		HouseholdID: householdID,
		Year:        year,
		Month:       month,
	})
	if err != nil {
		writeErrorFromSettlementDomain(w, err)
		return
	}

	members := make([]map[string]any, 0, len(out.Members))
	for _, m := range out.Members {
		members = append(members, map[string]any{
			"member_id":         m.MemberID,
			"name":              m.Name,
			"paid_cents":        m.PaidCents,
			"expected_share":    m.ExpectedShare,
			"net_balance_cents": m.NetBalanceCents,
			"salary_weight_bps": m.SalaryWeightBps,
		})
	}

	transfers := make([]map[string]any, 0, len(out.Transfers))
	for _, t := range out.Transfers {
		transfers = append(transfers, map[string]any{
			"from_member_id": t.FromMemberID,
			"to_member_id":   t.ToMemberID,
			"amount_cents":   t.AmountCents,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"household_id":              out.HouseholdID,
			"year":                      out.Year,
			"month":                     out.Month,
			"settlement_mode":           out.SettlementMode,
			"effective_settlement_mode": out.EffectiveSettlementMode,
			"fallback_reason":           out.FallbackReason,
			"total_shared_cents":        out.TotalSharedCents,
			"included_expense_count":    out.IncludedExpenseCount,
			"excluded_voucher_count":    out.ExcludedVoucherCount,
			"members":                   members,
			"transfers":                 transfers,
		},
	})
}

func writeErrorFromSettlementDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "the requested resource was not found")
	case errors.Is(err, settlement.ErrNoMembers):
		writeError(w, http.StatusBadRequest, "NO_MEMBERS", "household must have at least one member")
	case isValidationError(err):
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		slog.Error("settlement handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}

func isValidationError(err error) bool {
	return strings.Contains(err.Error(), "household_id is required") ||
		strings.Contains(err.Error(), "month must be between 1 and 12") ||
		strings.Contains(err.Error(), "year is out of range")
}
