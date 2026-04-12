package httpadapter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "micha/backend/internal/adapters/http"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/inbound"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	t.Parallel()

	registerUC := &mockRegisterUser{
		returnOutput: inbound.RegisterUserOutput{UserID: "user-123"},
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: registerUC,
			Login:    &mockLogin{},
		},
		Expense:        httpadapter.ExpenseHandlerDeps{},
		Household:      httpadapter.HouseholdHandlerDeps{},
		Member:         httpadapter.MemberHandlerDeps{},
		Settlement:     httpadapter.SettlementHandlerDeps{},
		Category:       httpadapter.CategoryHandlerDeps{},
		SplitConfig:    httpadapter.SplitConfigHandlerDeps{},
		JWTValidator:   &mockTokenValidator{},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "POST", "/v1/auth/register", map[string]string{
		"email":    "test@example.com",
		"password": "SecurePass123!",
	})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusCreated)
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["user_id"] != "user-123" {
		t.Errorf("user_id = %v; want user-123", data["user_id"])
	}

	// Verify input was passed correctly
	if registerUC.lastInput.Email != "test@example.com" {
		t.Errorf("input email = %q; want test@example.com", registerUC.lastInput.Email)
	}
}

func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	t.Parallel()

	registerUC := &mockRegisterUser{
		returnErr: user.ErrInvalidEmail,
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: registerUC,
			Login:    &mockLogin{},
		},
		JWTValidator:   &mockTokenValidator{},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "POST", "/v1/auth/register", map[string]string{
		"email":    "invalid-email",
		"password": "SecurePass123!",
	})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "INVALID_EMAIL" {
		t.Errorf("code = %v; want INVALID_EMAIL", errObj["code"])
	}
}

func TestAuthHandler_Register_WeakPassword(t *testing.T) {
	t.Parallel()

	registerUC := &mockRegisterUser{
		returnErr: user.ErrWeakPassword,
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: registerUC,
			Login:    &mockLogin{},
		},
		JWTValidator:   &mockTokenValidator{},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "POST", "/v1/auth/register", map[string]string{
		"email":    "test@example.com",
		"password": "weak",
	})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "WEAK_PASSWORD" {
		t.Errorf("code = %v; want WEAK_PASSWORD", errObj["code"])
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	t.Parallel()

	registerUC := &mockRegisterUser{
		returnErr: shared.ErrAlreadyExists,
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: registerUC,
			Login:    &mockLogin{},
		},
		JWTValidator:   &mockTokenValidator{},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "POST", "/v1/auth/register", map[string]string{
		"email":    "existing@example.com",
		"password": "SecurePass123!",
	})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusConflict)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "EMAIL_TAKEN" {
		t.Errorf("code = %v; want EMAIL_TAKEN", errObj["code"])
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	t.Parallel()

	loginUC := &mockLogin{
		returnOutput: inbound.LoginOutput{Token: "jwt-token-123"},
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    loginUC,
		},
		JWTValidator:   &mockTokenValidator{},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "POST", "/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "SecurePass123!",
	})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["token"] != "jwt-token-123" {
		t.Errorf("token = %v; want jwt-token-123", data["token"])
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	t.Parallel()

	loginUC := &mockLogin{
		returnErr: shared.ErrInvalidCredentials,
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    loginUC,
		},
		JWTValidator:   &mockTokenValidator{},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "POST", "/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusUnauthorized)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "INVALID_CREDENTIALS" {
		t.Errorf("code = %v; want INVALID_CREDENTIALS", errObj["code"])
	}
}

func TestAuthHandler_Me_Success(t *testing.T) {
	t.Parallel()

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		JWTValidator:   validator,
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}
	if data["user_id"] != "user-123" {
		t.Errorf("user_id = %v; want user-123", data["user_id"])
	}
	if data["email"] != "test@example.com" {
		t.Errorf("email = %v; want test@example.com", data["email"])
	}
}

func TestAuthHandler_Me_OwnerSessionIncludesSelectableMembers(t *testing.T) {
	t.Parallel()

	repo := newMockMemberRepo()
	now := time.Now()
	owner, err := member.NewFromAttributes(member.Attributes{
		ID:          member.ID("member-owner"),
		HouseholdID: "household-1",
		Name:        "Owner",
		Email:       "owner@example.com",
		UserID:      "user-owner",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("owner seed: %v", err)
	}
	other, err := member.NewFromAttributes(member.Attributes{
		ID:          member.ID("member-other"),
		HouseholdID: "household-1",
		Name:        "Other",
		Email:       "other@example.com",
		CreatedAt:   now.Add(time.Minute),
		UpdatedAt:   now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("other seed: %v", err)
	}
	repo.Save(context.Background(), owner)
	repo.Save(context.Background(), other)

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
			Members:  repo,
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-owner", returnEmail: "owner@example.com"},
		MemberRepo:     repo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/auth/me?household_id=household-1", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}
	session, ok := data["session"].(map[string]any)
	if !ok {
		t.Fatal("expected session object in response")
	}
	selectable, ok := session["selectable_members"].([]any)
	if !ok {
		t.Fatal("expected selectable_members array in session")
	}
	if len(selectable) != 2 {
		t.Fatalf("selectable_members len = %d; want 2", len(selectable))
	}
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	t.Parallel()

	validator := &mockTokenValidator{
		returnErr: shared.ErrInvalidCredentials,
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		JWTValidator:   validator,
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Me_MissingAuth(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		JWTValidator:   &mockTokenValidator{},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/auth/me", nil)
	// No Authorization header
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusUnauthorized)
	}
}
