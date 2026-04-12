package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

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
		HouseholdID       string `json:"household_id"`
		PaidByMemberID    string `json:"paid_by_member_id"`
		AmountCents       int64  `json:"amount_cents"`
		Description       string `json:"description"`
		IsShared          *bool  `json:"is_shared"`
		Currency          string `json:"currency"`
		PaymentMethod     string `json:"payment_method"`
		ExpenseType       string `json:"expense_type"`
		CardID            string `json:"card_id"`
		CardName          string `json:"card_name"`
		Category          string `json:"category"`
		CategoryID        string `json:"category_id"`
		TotalInstallments int    `json:"total_installments"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	isShared := true
	if body.IsShared != nil {
		isShared = *body.IsShared
	}

	currency := body.Currency
	if currency == "" {
		currency = "MXN"
	}

	categoryID := body.CategoryID
	if categoryID == "" {
		categoryID = body.Category
	}

	currentUserID, _ := UserIDFromContext(r.Context())

	input := inbound.RegisterExpenseInput{
		HouseholdID:       body.HouseholdID,
		PaidByMemberID:    body.PaidByMemberID,
		CurrentUserID:     currentUserID,
		AmountCents:       body.AmountCents,
		Description:       body.Description,
		IsShared:          isShared,
		Currency:          currency,
		PaymentMethod:     body.PaymentMethod,
		ExpenseType:       body.ExpenseType,
		CardID:            body.CardID,
		CardName:          body.CardName,
		CategoryID:        categoryID,
		TotalInstallments: body.TotalInstallments,
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
		"id":                 string(attrs.ID),
		"household_id":       attrs.HouseholdID,
		"paid_by_member_id":  attrs.PaidByMemberID,
		"amount_cents":       attrs.AmountCents,
		"description":        attrs.Description,
		"is_shared":          attrs.IsShared,
		"currency":           attrs.Currency,
		"payment_method":     string(attrs.PaymentMethod),
		"expense_type":       string(attrs.ExpenseType),
		"card_id":            attrs.CardID,
		"card_name":          attrs.CardName,
		"category_id":        attrs.CategoryID,
		"total_installments": attrs.TotalInstallments,
		"created_at":         attrs.CreatedAt,
		"updated_at":         attrs.UpdatedAt,
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

// writeErrorFromDomain maps a domain/use-case error to an HTTP response.
func writeErrorFromDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "the requested resource was not found")
	case errors.Is(err, shared.ErrInvalidMoney):
		writeError(w, http.StatusBadRequest, "INVALID_MONEY", "amount_cents must be greater than zero")
	case errors.Is(err, expense.ErrInvalidHouseholdID):
		writeError(w, http.StatusBadRequest, "INVALID_HOUSEHOLD_ID", "household_id is required")
	case errors.Is(err, expense.ErrInvalidPaidByMemberID):
		writeError(w, http.StatusBadRequest, "INVALID_PAID_BY_MEMBER_ID", "paid_by_member_id is required")
	case errors.Is(err, expense.ErrInvalidCurrency):
		writeError(w, http.StatusBadRequest, "INVALID_CURRENCY", "currency must be a 3-letter code")
	case errors.Is(err, expense.ErrInvalidPaymentMethod):
		writeError(w, http.StatusBadRequest, "INVALID_PAYMENT_METHOD", "payment_method must be cash, card, transfer or voucher")
	case errors.Is(err, expense.ErrInvalidExpenseType):
		writeError(w, http.StatusBadRequest, "INVALID_EXPENSE_TYPE", "expense_type must be fixed, variable or msi")
	case errors.Is(err, expense.ErrInvalidCategory):
		writeError(w, http.StatusBadRequest, "INVALID_CATEGORY", "category must be rent, auto, streaming, food, personal, savings or other")
	case errors.Is(err, shared.ErrAlreadyDeleted):
		writeError(w, http.StatusBadRequest, "ALREADY_DELETED", "expense has already been deleted")
	case errors.Is(err, shared.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "you are not allowed to register expenses with this member")
	default:
		slog.Error("expense handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
