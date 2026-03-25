package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// RecurringExpenseHandlerDeps groups all use case dependencies for the recurring expense resource.
type RecurringExpenseHandlerDeps struct {
	Create   inbound.CreateRecurringExpenseUseCase
	Get      inbound.GetRecurringExpenseUseCase
	List     inbound.ListRecurringExpensesUseCase
	Update   inbound.UpdateRecurringExpenseUseCase
	Delete   inbound.DeleteRecurringExpenseUseCase
	Generate inbound.GenerateRecurringExpensesUseCase
}

// recurringExpenseHandler handles HTTP requests for the recurring expense resource.
type recurringExpenseHandler struct {
	deps RecurringExpenseHandlerDeps
}

func newRecurringExpenseHandler(deps RecurringExpenseHandlerDeps) recurringExpenseHandler {
	return recurringExpenseHandler{deps: deps}
}

// handleCreate handles POST /v1/recurring-expenses.
func (h recurringExpenseHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		HouseholdID       string  `json:"household_id"`
		PaidByMemberID    string  `json:"paid_by_member_id"`
		AmountCents       int64   `json:"amount_cents"`
		Description       string  `json:"description"`
		CategoryID        string  `json:"category_id"`
		ExpenseType       string  `json:"expense_type"`
		RecurrencePattern string  `json:"recurrence_pattern"`
		StartDate         string  `json:"start_date"`
		EndDate           *string `json:"end_date"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	startDate, err := time.Parse(time.DateOnly, body.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_START_DATE", "start_date must be in YYYY-MM-DD format")
		return
	}

	var endDate *time.Time
	if body.EndDate != nil {
		parsed, err := time.Parse(time.DateOnly, *body.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_END_DATE", "end_date must be in YYYY-MM-DD format")
			return
		}
		endDate = &parsed
	}

	input := inbound.CreateRecurringExpenseInput{
		HouseholdID:       body.HouseholdID,
		PaidByMemberID:    body.PaidByMemberID,
		AmountCents:       body.AmountCents,
		Description:       body.Description,
		CategoryID:        body.CategoryID,
		ExpenseType:       body.ExpenseType,
		RecurrencePattern: body.RecurrencePattern,
		StartDate:         startDate,
		EndDate:           endDate,
	}

	out, err := h.deps.Create.Execute(r.Context(), input)
	if err != nil {
		writeRecurringExpenseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]string{"recurring_expense_id": out.RecurringExpenseID},
	})
}

// handleGet handles GET /v1/recurring-expenses/{id}.
func (h recurringExpenseHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseRecurringExpenseID(w, r)
	if !ok {
		return
	}

	re, err := h.deps.Get.Execute(r.Context(), id)
	if err != nil {
		writeRecurringExpenseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": recurringExpenseJSON(re)})
}

// handleList handles GET /v1/recurring-expenses?household_id=&limit=&offset=.
func (h recurringExpenseHandler) handleList(w http.ResponseWriter, r *http.Request) {
	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "household_id query parameter is required")
		return
	}

	limit := queryInt(r, "limit", 100)
	offset := queryInt(r, "offset", 0)

	recurringExpenses, err := h.deps.List.Execute(r.Context(), inbound.ListRecurringExpensesQuery{
		HouseholdID: householdID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeRecurringExpenseError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(recurringExpenses))
	for _, re := range recurringExpenses {
		items = append(items, recurringExpenseJSON(re))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// handleUpdate handles PATCH /v1/recurring-expenses/{id}.
func (h recurringExpenseHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseRecurringExpenseID(w, r)
	if !ok {
		return
	}

	var body struct {
		Description *string `json:"description"`
		AmountCents *int64  `json:"amount_cents"`
		CategoryID  *string `json:"category_id"`
		IsActive    *bool   `json:"is_active"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	re, err := h.deps.Update.Execute(r.Context(), inbound.UpdateRecurringExpenseCommand{
		ID:          id,
		Description: body.Description,
		AmountCents: body.AmountCents,
		CategoryID:  body.CategoryID,
		IsActive:    body.IsActive,
	})
	if err != nil {
		writeRecurringExpenseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": recurringExpenseJSON(re)})
}

// handleDelete handles DELETE /v1/recurring-expenses/{id}.
func (h recurringExpenseHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseRecurringExpenseID(w, r)
	if !ok {
		return
	}

	if err := h.deps.Delete.Execute(r.Context(), id); err != nil {
		writeRecurringExpenseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGenerate handles POST /v1/recurring-expenses/generate.
func (h recurringExpenseHandler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		HouseholdID string  `json:"household_id"`
		AsOfDate    *string `json:"as_of_date"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	var asOfDate time.Time
	if body.AsOfDate != nil {
		parsed, err := time.Parse(time.DateOnly, *body.AsOfDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_AS_OF_DATE", "as_of_date must be in YYYY-MM-DD format")
			return
		}
		asOfDate = parsed
	} else {
		asOfDate = time.Now()
	}

	out, err := h.deps.Generate.Execute(r.Context(), inbound.GenerateRecurringExpensesCommand{
		HouseholdID: body.HouseholdID,
		AsOfDate:    asOfDate,
	})
	if err != nil {
		writeRecurringExpenseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"generated_count": out.GeneratedCount,
			"expense_ids":     out.ExpenseIDs,
		},
	})
}

// recurringExpenseJSON converts a RecurringExpense to a JSON-serialisable map.
func recurringExpenseJSON(re recurringexpense.RecurringExpense) map[string]any {
	attrs := re.Attributes()
	m := map[string]any{
		"id":                   string(attrs.ID),
		"household_id":         attrs.HouseholdID,
		"paid_by_member_id":    attrs.PaidByMemberID,
		"amount_cents":         attrs.AmountCents,
		"description":          attrs.Description,
		"category_id":          attrs.CategoryID,
		"expense_type":         string(attrs.ExpenseType),
		"recurrence_pattern":   string(attrs.RecurrencePattern),
		"start_date":           attrs.StartDate.Format(time.DateOnly),
		"next_generation_date": attrs.NextGenerationDate.Format(time.DateOnly),
		"is_active":            attrs.IsActive,
		"created_at":           attrs.CreatedAt,
		"updated_at":           attrs.UpdatedAt,
	}
	if attrs.EndDate != nil {
		m["end_date"] = attrs.EndDate.Format(time.DateOnly)
	}
	if attrs.DeletedAt != nil {
		m["deleted_at"] = attrs.DeletedAt
	}
	return m
}

// parseRecurringExpenseID extracts and validates the {id} path value as a UUID.
func parseRecurringExpenseID(w http.ResponseWriter, r *http.Request) (string, bool) {
	raw := r.PathValue("id")
	if _, err := uuid.Parse(raw); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID")
		return "", false
	}
	return raw, true
}

// writeRecurringExpenseError maps a domain/use-case error to an HTTP response.
func writeRecurringExpenseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "the requested resource was not found")
	case errors.Is(err, shared.ErrInvalidMoney):
		writeError(w, http.StatusBadRequest, "INVALID_MONEY", "amount_cents must be greater than zero")
	case errors.Is(err, recurringexpense.ErrInvalidHouseholdID):
		writeError(w, http.StatusBadRequest, "INVALID_HOUSEHOLD_ID", "household_id is required")
	case errors.Is(err, recurringexpense.ErrInvalidPaidByMemberID):
		writeError(w, http.StatusBadRequest, "INVALID_PAID_BY_MEMBER_ID", "paid_by_member_id is required")
	case errors.Is(err, recurringexpense.ErrInvalidRecurrencePattern):
		writeError(w, http.StatusBadRequest, "INVALID_RECURRENCE_PATTERN", "recurrence_pattern must be monthly, biweekly or weekly")
	case errors.Is(err, recurringexpense.ErrInvalidDateRange):
		writeError(w, http.StatusBadRequest, "INVALID_DATE_RANGE", "end_date must be after start_date")
	case errors.Is(err, recurringexpense.ErrInvalidNextGenerationDate):
		writeError(w, http.StatusBadRequest, "INVALID_NEXT_GENERATION_DATE", "next_generation_date cannot be before start_date")
	case errors.Is(err, shared.ErrAlreadyDeleted):
		writeError(w, http.StatusBadRequest, "ALREADY_DELETED", "recurring expense has already been deleted")
	default:
		slog.Error("recurring expense handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
