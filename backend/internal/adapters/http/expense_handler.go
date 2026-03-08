package httpadapter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

const maxRequestBytes = 1 << 20 // 1 MB

// ExpenseHandlerDeps groups all use case dependencies for the expense resource.
type ExpenseHandlerDeps struct {
	Register inbound.RegisterExpenseUseCase
	Get      inbound.GetExpenseUseCase
	List     inbound.ListExpensesUseCase
	Patch    inbound.PatchExpenseUseCase
	Delete   inbound.DeleteExpenseUseCase
}

// expenseHandler handles HTTP requests for the expense resource.
type expenseHandler struct {
	deps ExpenseHandlerDeps
}

func newExpenseHandler(deps ExpenseHandlerDeps) expenseHandler {
	return expenseHandler{deps: deps}
}

// handleCreate handles POST /v1/expenses.
func (h expenseHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		HouseholdID string `json:"household_id"`
		AmountCents int64  `json:"amount_cents"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	input := inbound.RegisterExpenseInput{
		HouseholdID: body.HouseholdID,
		AmountCents: body.AmountCents,
		Description: body.Description,
	}

	out, err := h.deps.Register.Execute(r.Context(), input)
	if err != nil {
		writeErrorFromDomain(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]string{"expense_id": out.ExpenseID},
	})
}

// handleGet handles GET /v1/expenses/{id}.
func (h expenseHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, ok := parseExpenseID(w, r)
	if !ok {
		return
	}

	e, err := h.deps.Get.Execute(r.Context(), id)
	if err != nil {
		writeErrorFromDomain(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": expenseJSON(e)})
}

// handleList handles GET /v1/expenses?household_id=&limit=&offset=.
func (h expenseHandler) handleList(w http.ResponseWriter, r *http.Request) {
	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "household_id query parameter is required")
		return
	}

	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	expenses, err := h.deps.List.Execute(r.Context(), inbound.ListExpensesQuery{
		HouseholdID: householdID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		writeErrorFromDomain(w, err)
		return
	}

	items := make([]map[string]any, 0, len(expenses))
	for _, e := range expenses {
		items = append(items, expenseJSON(e))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// handlePatch handles PATCH /v1/expenses/{id}.
func (h expenseHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	id, ok := parseExpenseID(w, r)
	if !ok {
		return
	}

	var body struct {
		Description *string `json:"description"`
		AmountCents *int64  `json:"amount_cents"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	e, err := h.deps.Patch.Execute(r.Context(), inbound.PatchExpenseCommand{
		ID:          id,
		Description: body.Description,
		AmountCents: body.AmountCents,
	})
	if err != nil {
		writeErrorFromDomain(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": expenseJSON(e)})
}

// handleDelete handles DELETE /v1/expenses/{id}.
func (h expenseHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseExpenseID(w, r)
	if !ok {
		return
	}

	if err := h.deps.Delete.Execute(r.Context(), id); err != nil {
		writeErrorFromDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// expenseJSON converts an Expense to a JSON-serialisable map.
func expenseJSON(e expense.Expense) map[string]any {
	attrs := e.Attributes()
	m := map[string]any{
		"id":           string(attrs.ID),
		"household_id": attrs.HouseholdID,
		"amount_cents": attrs.AmountCents,
		"description":  attrs.Description,
		"created_at":   attrs.CreatedAt,
		"updated_at":   attrs.UpdatedAt,
	}
	if attrs.DeletedAt != nil {
		m["deleted_at"] = attrs.DeletedAt
	}
	return m
}

// parseExpenseID extracts and validates the {id} path value as a UUID.
func parseExpenseID(w http.ResponseWriter, r *http.Request) (string, bool) {
	raw := r.PathValue("id")
	if _, err := uuid.Parse(raw); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "id must be a valid UUID")
		return "", false
	}
	return raw, true
}

// decodeJSON decodes the request body into dst, enforcing size limit and unknown-field rejection.
// Returns non-nil error and writes a 400 response when decoding fails.
func decodeJSON(r *http.Request, w http.ResponseWriter, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "request body is invalid: "+err.Error())
		return err
	}
	return nil
}

// queryInt reads an integer query param, returning fallback on parse failure.
func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// writeJSON serialises data to JSON and writes it with status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("expense handler: failed to encode response", "error", err)
	}
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// writeErrorFromDomain maps a domain/use-case error to an HTTP response.
func writeErrorFromDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "the requested resource was not found")
	case errors.Is(err, shared.ErrInvalidMoney):
		writeError(w, http.StatusBadRequest, "INVALID_MONEY", "amount_cents must be greater than zero")
	case errors.Is(err, shared.ErrAlreadyDeleted):
		writeError(w, http.StatusBadRequest, "ALREADY_DELETED", "expense has already been deleted")
	default:
		slog.Error("expense handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
