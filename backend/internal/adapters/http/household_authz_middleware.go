package httpadapter

import (
	"context"
	"net/http"

	"micha/backend/internal/ports/outbound"
)

const (
	contextKeyMemberID    contextKey = "member_id"
	contextKeyHouseholdID contextKey = "household_id"
)

// MemberIDFromContext extracts the caller's member ID (scoped to the current household) from context.
// Returns ("", false) if absent.
func MemberIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKeyMemberID).(string)
	return v, ok && v != ""
}

// HouseholdIDFromContext extracts the current household ID from context.
// Returns ("", false) if absent.
func HouseholdIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKeyHouseholdID).(string)
	return v, ok && v != ""
}

// HouseholdAuthzMiddleware verifies that the authenticated user is a member of the household
// identified by the {household_id} path parameter. On success it injects both the household ID
// and the caller's member ID into the request context.
func HouseholdAuthzMiddleware(memberRepo outbound.MemberRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
				return
			}

			householdID := r.PathValue("household_id")
			if householdID == "" {
				writeError(w, http.StatusBadRequest, "BAD_REQUEST", "household_id path parameter is required")
				return
			}

			m, err := memberRepo.FindByUserID(r.Context(), householdID, userID)
			if err != nil {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "you are not a member of this household")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyHouseholdID, householdID)
			ctx = context.WithValue(ctx, contextKeyMemberID, string(m.ID()))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
