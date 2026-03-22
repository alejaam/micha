package httpadapter

import (
	"context"
	"net/http"
	"strings"

	"micha/backend/internal/ports/outbound"
)

type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyEmail  contextKey = "email"
)

// UserIDFromContext extracts the authenticated user ID from the request context.
// Returns ("", false) if the value is absent or not a non-empty string.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKeyUserID).(string)
	return v, ok && v != ""
}

// EmailFromContext extracts the authenticated user email from the request context.
// Returns ("", false) if the value is absent or not a non-empty string.
func EmailFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(contextKeyEmail).(string)
	return v, ok && v != ""
}

// AuthMiddleware validates the Bearer JWT on protected routes.
func AuthMiddleware(validator outbound.TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			userID, email, err := validator.Validate(tokenStr)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			ctx = context.WithValue(ctx, contextKeyEmail, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
