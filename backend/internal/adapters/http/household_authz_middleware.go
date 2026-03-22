package httpadapter

import (
	"context"
	"errors"
	"net/http"

	"micha/backend/internal/domain/shared"
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

// HouseholdMemberOrEmptyAuthzMiddleware allows access when either:
// 1) the authenticated user is already a member of the household, or
// 2) the household has no members yet (bootstrap onboarding flow).
//
// This is intended for member creation so the first member can be added right
// after household creation.
func HouseholdMemberOrEmptyAuthzMiddleware(memberRepo outbound.MemberRepository) func(http.Handler) http.Handler {
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
			if err == nil {
				ctx := context.WithValue(r.Context(), contextKeyHouseholdID, householdID)
				ctx = context.WithValue(ctx, contextKeyMemberID, string(m.ID()))
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !errors.Is(err, shared.ErrNotFound) {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
				return
			}

			members, listErr := memberRepo.ListByHousehold(r.Context(), householdID, 1, 0)
			if listErr != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
				return
			}

			if len(members) > 0 {
				writeError(w, http.StatusForbidden, "FORBIDDEN", "you are not a member of this household")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyHouseholdID, householdID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
