package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"micha/backend/internal/domain/member"
	"micha/backend/internal/ports/inbound"
)

// MemberHandlerDeps groups all use case dependencies for the member resource.
type MemberHandlerDeps struct {
	Register inbound.RegisterMemberUseCase
	List     inbound.ListMembersUseCase
}

// memberHandler handles HTTP requests for the member resource.
type memberHandler struct {
	deps MemberHandlerDeps
}

func newMemberHandler(deps MemberHandlerDeps) memberHandler {
	return memberHandler{deps: deps}
}

// handleCreate handles POST /v1/households/{household_id}/members.
// If the authenticated user's email matches the new member's email, the user_id
// is automatically linked to the member.
func (h memberHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
		return
	}

	var body struct {
		Name               string `json:"name"`
		Email              string `json:"email"`
		MonthlySalaryCents int64  `json:"monthly_salary_cents"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	// Pass caller identity so the use case can apply the auto-link rule.
	userID, _ := UserIDFromContext(r.Context())
	authEmail, _ := EmailFromContext(r.Context())

	out, err := h.deps.Register.Execute(r.Context(), inbound.RegisterMemberInput{
		HouseholdID:        householdID,
		Name:               body.Name,
		Email:              body.Email,
		MonthlySalaryCents: body.MonthlySalaryCents,
		CallerUserID:       userID,
		CallerEmail:        authEmail,
	})
	if err != nil {
		writeErrorFromMemberDomain(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]string{"member_id": out.MemberID},
	})
}

// handleList handles GET /v1/households/{household_id}/members?limit=&offset=.
func (h memberHandler) handleList(w http.ResponseWriter, r *http.Request) {
	householdID, ok := parseHouseholdID(w, r)
	if !ok {
		return
	}

	members, err := h.deps.List.Execute(r.Context(), inbound.ListMembersQuery{
		HouseholdID: householdID,
		Limit:       queryInt(r, "limit", 20),
		Offset:      queryInt(r, "offset", 0),
	})
	if err != nil {
		writeErrorFromMemberDomain(w, err)
		return
	}

	items := make([]map[string]any, 0, len(members))
	for _, item := range members {
		attrs := item.Attributes()
		items = append(items, map[string]any{
			"id":                   string(attrs.ID),
			"household_id":         attrs.HouseholdID,
			"name":                 attrs.Name,
			"email":                attrs.Email,
			"monthly_salary_cents": attrs.MonthlySalaryCents,
			"user_id":              attrs.UserID,
			"created_at":           attrs.CreatedAt,
			"updated_at":           attrs.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func parseHouseholdID(w http.ResponseWriter, r *http.Request) (string, bool) {
	raw := r.PathValue("household_id")
	if _, err := uuid.Parse(raw); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HOUSEHOLD_ID", "household_id must be a valid UUID")
		return "", false
	}
	return raw, true
}

func writeErrorFromMemberDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, member.ErrInvalidName):
		writeError(w, http.StatusBadRequest, "INVALID_NAME", "member name is required")
	case errors.Is(err, member.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", "email is invalid")
	case errors.Is(err, member.ErrInvalidSalary):
		writeError(w, http.StatusBadRequest, "INVALID_SALARY", "monthly_salary_cents must be greater than or equal to zero")
	default:
		slog.Error("member handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
