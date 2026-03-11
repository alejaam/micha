package httpadapter

import (
	"errors"
	"log/slog"
	"net/http"

	authapp "micha/backend/internal/application/auth"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// AuthHandlerDeps groups use case dependencies for the auth resource.
type AuthHandlerDeps struct {
	Register inbound.RegisterUserUseCase
	Login    inbound.LoginUseCase
}

type authHandler struct{ deps AuthHandlerDeps }

func newAuthHandler(deps AuthHandlerDeps) authHandler { return authHandler{deps: deps} }

// handleRegister handles POST /v1/auth/register.
func (h authHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	out, err := h.deps.Register.Execute(r.Context(), inbound.RegisterUserInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]string{"user_id": out.UserID}})
}

// handleLogin handles POST /v1/auth/login.
func (h authHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, w, &body); err != nil {
		return
	}

	out, err := h.deps.Login.Execute(r.Context(), inbound.LoginInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"token": out.Token}})
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authapp.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect")
	case errors.Is(err, shared.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "EMAIL_TAKEN", "an account with this email already exists")
	default:
		slog.Error("auth handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
