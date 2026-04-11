package httpadapter

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// AuthHandlerDeps groups use case dependencies for the auth resource.
type AuthHandlerDeps struct {
	Register inbound.RegisterUserUseCase
	Login    inbound.LoginUseCase
	Members  outbound.MemberRepository
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

// handleMe handles GET /v1/auth/me — returns the authenticated user's profile
// extracted from the JWT claims injected by AuthMiddleware.
func (h authHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or invalid authorization header")
		return
	}

	email, _ := EmailFromContext(r.Context())
	householdID := strings.TrimSpace(r.URL.Query().Get("household_id"))

	response := map[string]any{
		"user_id": userID,
		"email":   email,
	}

	if h.deps.Members != nil {
		households, err := h.deps.Members.ListHouseholdIDsByUserID(r.Context(), userID)
		if err == nil {
			roles := make([]map[string]any, 0, len(households))
			for _, hhID := range households {
				role, roleErr := resolveSessionRole(r.Context(), h.deps.Members, hhID, userID)
				if roleErr != nil {
					continue
				}
				roles = append(roles, map[string]any{
					"household_id": hhID,
					"role":         role,
					"permissions":  permissionsByRole(role),
				})
			}

			if householdID != "" {
				for _, item := range roles {
					if item["household_id"] == householdID {
						response["session"] = item
						break
					}
				}
			} else {
				response["households"] = roles
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": response,
	})
}

func resolveSessionRole(ctx context.Context, repo outbound.MemberRepository, householdID, userID string) (string, error) {
	members, err := repo.ListAllByHousehold(ctx, householdID)
	if err != nil {
		return "", err
	}
	if len(members) == 0 {
		return "", shared.ErrNotFound
	}

	ownerUserID := resolveOwnerUserIDForSession(members)
	if ownerUserID == userID {
		return "owner", nil
	}

	_, err = repo.FindByUserID(ctx, householdID, userID)
	if err != nil {
		return "", err
	}

	return "member", nil
}

func resolveOwnerUserIDForSession(members []member.Member) string {
	linked := make([]member.Member, 0, len(members))
	for _, m := range members {
		if strings.TrimSpace(m.UserID()) != "" {
			linked = append(linked, m)
		}
	}
	if len(linked) == 0 {
		return ""
	}

	sort.SliceStable(linked, func(i, j int) bool {
		return linked[i].CreatedAt().Before(linked[j].CreatedAt())
	})

	return linked[0].UserID()
}

func permissionsByRole(role string) []string {
	if role == "owner" {
		return []string{
			"member.invite",
			"expense.create.fixed",
			"expense.create.on_behalf_temp",
			"settlement.read",
		}
	}

	return []string{
		"expense.create.msi",
		"expense.create.occasional",
		"settlement.read",
	}
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shared.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect")
	case errors.Is(err, shared.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "EMAIL_TAKEN", "an account with this email already exists")
	case errors.Is(err, user.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", "email address is invalid")
	case errors.Is(err, user.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", "password is too weak")
	default:
		slog.Error("auth handler: internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
	}
}
