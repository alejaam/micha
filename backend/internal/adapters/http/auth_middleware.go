package httpadapter

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyEmail  contextKey = "email"
)

// TokenValidator validates a Bearer JWT and returns the embedded claims.
type TokenValidator interface {
	Validate(tokenString string) (userID, email string, err error)
}

// AuthMiddleware validates the Bearer JWT on protected routes.
func AuthMiddleware(validator TokenValidator) func(http.Handler) http.Handler {
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
