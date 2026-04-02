package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	"micha/backend/internal/domain/card"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// CardHandlerDeps groups all use case dependencies for the card resource.
type CardHandlerDeps struct {
	Register inbound.RegisterCardUseCase
	List     inbound.ListCardsUseCase
	Delete   inbound.DeleteCardUseCase
}

// cardHandler handles HTTP requests for the card resource.
type cardHandler struct {
	deps CardHandlerDeps
}

func newCardHandler(deps CardHandlerDeps) cardHandler {
	return cardHandler{deps: deps}
}

// handleCreate handles POST /v1/households/{household_id}/cards.
func (h cardHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
		return
	}

	var body struct {
		BankName  string `json:"bank_name"`
		CardName  string `json:"card_name"`
		CutoffDay int    `json:"cutoff_day"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	out, err := h.deps.Register.Execute(r.Context(), inbound.RegisterCardInput{
		HouseholdID: householdID,
		BankName:    body.BankName,
		CardName:    body.CardName,
		CutoffDay:   body.CutoffDay,
	})
	if err != nil {
		writeErrorFromCardDomain(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]string{"card_id": out.CardID},
	})
}

// handleList handles GET /v1/households/{household_id}/cards.
func (h cardHandler) handleList(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
		return
	}

	cards, err := h.deps.List.Execute(r.Context(), inbound.ListCardsQuery{
		HouseholdID: householdID,
	})
	if err != nil {
		writeErrorFromCardDomain(w, err)
		return
	}

	items := make([]map[string]any, 0, len(cards))
	for _, item := range cards {
		attrs := item.Attributes()
		items = append(items, map[string]any{
			"id":           string(attrs.ID),
			"household_id": attrs.HouseholdID,
			"bank_name":    attrs.BankName,
			"card_name":    attrs.CardName,
			"cutoff_day":   attrs.CutoffDay,
			"created_at":   attrs.CreatedAt,
			"updated_at":   attrs.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// handleDelete handles DELETE /v1/cards/{card_id}?household_id=.
func (h cardHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("card_id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_CARD_ID", "card_id is required")
		return
	}

	householdID := r.URL.Query().Get("household_id")
	if householdID == "" {
		writeError(w, http.StatusBadRequest, "INVALID_HOUSEHOLD_ID", "household_id query param is required")
		return
	}

	err := h.deps.Delete.Execute(r.Context(), inbound.DeleteCardInput{
		CardID:      cardID,
		HouseholdID: householdID,
	})
	if err != nil {
		writeErrorFromCardDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeErrorFromCardDomain maps card domain errors to HTTP responses.
func writeErrorFromCardDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "card not found")
	case errors.Is(err, shared.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "ALREADY_EXISTS", "card already exists")
	case errors.Is(err, card.ErrInvalidBankName):
		writeError(w, http.StatusBadRequest, "INVALID_BANK_NAME", "invalid bank name")
	case errors.Is(err, card.ErrInvalidCardName):
		writeError(w, http.StatusBadRequest, "INVALID_CARD_NAME", "invalid card name")
	case errors.Is(err, card.ErrInvalidCutoffDay):
		writeError(w, http.StatusBadRequest, "INVALID_CUTOFF_DAY", "invalid cutoff day: must be between 1 and 31")
	case errors.Is(err, shared.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
	default:
		slog.Error("card handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
