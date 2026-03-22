package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	categoryapp "micha/backend/internal/application/category"
	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// CategoryHandlerDeps groups use case dependencies for the category resource.
type CategoryHandlerDeps struct {
	Create inbound.CreateCategoryUseCase
	List   inbound.ListCategoriesUseCase
	Delete inbound.DeleteCategoryUseCase
}

// SplitConfigHandlerDeps groups use case dependencies for the split config resource.
type SplitConfigHandlerDeps struct {
	Update inbound.UpdateSplitConfigUseCase
}

type categoryHandler struct {
	deps CategoryHandlerDeps
}

func newCategoryHandler(deps CategoryHandlerDeps) categoryHandler {
	return categoryHandler{deps: deps}
}

// handleCreate handles POST /v1/households/{household_id}/categories.
func (h categoryHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	householdID, ok := HouseholdIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var body struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	out, err := h.deps.Create.Execute(r.Context(), inbound.CreateCategoryInput{
		HouseholdID: householdID,
		Name:        body.Name,
		Slug:        body.Slug,
	})
	if err != nil {
		writeErrorFromCategoryDomain(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]string{"category_id": out.CategoryID},
	})
}

// handleList handles GET /v1/households/{household_id}/categories.
func (h categoryHandler) handleList(w http.ResponseWriter, r *http.Request) {
	householdID, ok := HouseholdIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	cats, err := h.deps.List.Execute(r.Context(), inbound.ListCategoriesQuery{HouseholdID: householdID})
	if err != nil {
		writeErrorFromCategoryDomain(w, err)
		return
	}

	items := make([]map[string]any, 0, len(cats))
	for _, c := range cats {
		attrs := c.Attributes()
		items = append(items, map[string]any{
			"id":           string(attrs.ID),
			"household_id": attrs.HouseholdID,
			"name":         attrs.Name,
			"slug":         attrs.Slug,
			"is_default":   c.IsDefault(),
			"created_at":   attrs.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

// handleDelete handles DELETE /v1/households/{household_id}/categories/{category_id}.
func (h categoryHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	householdID, ok := HouseholdIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	categoryID := r.PathValue("category_id")
	if categoryID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_CATEGORY_ID", "category_id path param is required")
		return
	}

	if err := h.deps.Delete.Execute(r.Context(), inbound.DeleteCategoryInput{
		HouseholdID: householdID,
		CategoryID:  categoryID,
	}); err != nil {
		writeErrorFromCategoryDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Split config handler ---

type splitConfigHandler struct {
	deps SplitConfigHandlerDeps
}

func newSplitConfigHandler(deps SplitConfigHandlerDeps) splitConfigHandler {
	return splitConfigHandler{deps: deps}
}

// handleUpdate handles PUT /v1/households/{household_id}/split-config.
func (h splitConfigHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	householdID, ok := HouseholdIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	var body struct {
		Splits []struct {
			MemberID   string `json:"member_id"`
			Percentage int    `json:"percentage"`
		} `json:"splits"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	splits := make([]household.MemberSplit, 0, len(body.Splits))
	for _, s := range body.Splits {
		splits = append(splits, household.MemberSplit{MemberID: s.MemberID, Percentage: float64(s.Percentage)})
	}

	if err := h.deps.Update.Execute(r.Context(), inbound.UpdateSplitConfigInput{
		HouseholdID: householdID,
		Splits:      splits,
	}); err != nil {
		writeErrorFromSplitConfigDomain(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- error mappers ---

func writeErrorFromCategoryDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, category.ErrInvalidSlug):
		writeError(w, http.StatusBadRequest, "INVALID_SLUG", "slug must be lowercase alphanumeric with hyphens, max 64 chars")
	case errors.Is(err, shared.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "ALREADY_EXISTS", "a category with this slug already exists")
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "category not found")
	case errors.Is(err, categoryapp.ErrCannotDeleteDefault):
		writeError(w, http.StatusForbidden, "CANNOT_DELETE_DEFAULT", "default categories cannot be deleted")
	default:
		slog.Error("category handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}

func writeErrorFromSplitConfigDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, household.ErrInvalidSplitConfig):
		writeError(w, http.StatusBadRequest, "INVALID_SPLIT_CONFIG", "split percentages must sum to 100")
	case errors.Is(err, household.ErrEmptySplitConfig):
		writeError(w, http.StatusBadRequest, "EMPTY_SPLIT_CONFIG", "split config must have at least one member")
	case errors.Is(err, shared.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "household not found")
	default:
		slog.Error("split config handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
