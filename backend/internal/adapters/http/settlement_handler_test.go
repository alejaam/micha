package httpadapter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "micha/backend/internal/adapters/http"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

func TestSettlementHandler_GetMonthly_Success(t *testing.T) {
	t.Parallel()

	calculateUC := &mockCalculateSettlement{
		returnOutput: makeTestSettlementOutput(),
	}

	memberRepo := newMockMemberRepo()
	memberRepo.seedMember("m-1", "550e8400-e29b-41d4-a716-446655440000", "user-123")

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: calculateUC,
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/550e8400-e29b-41d4-a716-446655440000/settlement?year=2026&month=3", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	resp := parseJSONResponse(t, rec)
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data object in response")
	}

	if data["total_shared_cents"].(float64) != 15000 {
		t.Errorf("total_shared_cents = %v; want 15000", data["total_shared_cents"])
	}

	transfers, ok := data["transfers"].([]any)
	if !ok || len(transfers) != 1 {
		t.Errorf("transfers = %v; want 1 transfer", transfers)
	}

	// Verify input was passed correctly
	if calculateUC.lastInput.Year != 2026 {
		t.Errorf("input year = %d; want 2026", calculateUC.lastInput.Year)
	}
	if calculateUC.lastInput.Month != 3 {
		t.Errorf("input month = %d; want 3", calculateUC.lastInput.Month)
	}
}

func TestSettlementHandler_GetMonthly_DefaultsToCurrentMonth(t *testing.T) {
	t.Parallel()

	calculateUC := &mockCalculateSettlement{
		returnOutput: inbound.CalculateSettlementOutput{
			HouseholdID:      "550e8400-e29b-41d4-a716-446655440000",
			TotalSharedCents: 0,
			SettlementMode:   household.SettlementModeEqual,
		},
	}

	memberRepo := newMockMemberRepo()
	memberRepo.seedMember("m-1", "550e8400-e29b-41d4-a716-446655440000", "user-123")

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: calculateUC,
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	// No year/month query params - should default to current
	req := makeJSONRequest(t, "GET", "/v1/households/550e8400-e29b-41d4-a716-446655440000/settlement", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
	}

	// Verify defaults were applied (should be current year/month)
	if calculateUC.lastInput.Year < 2020 {
		t.Errorf("input year = %d; should be current year", calculateUC.lastInput.Year)
	}
	if calculateUC.lastInput.Month < 1 || calculateUC.lastInput.Month > 12 {
		t.Errorf("input month = %d; should be 1-12", calculateUC.lastInput.Month)
	}
}

func TestSettlementHandler_GetMonthly_HouseholdNotFound(t *testing.T) {
	t.Parallel()

	calculateUC := &mockCalculateSettlement{
		returnErr: shared.ErrNotFound,
	}

	memberRepo := newMockMemberRepo()
	memberRepo.seedMember("m-1", "550e8400-e29b-41d4-a716-446655440000", "user-123")

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: calculateUC,
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/550e8400-e29b-41d4-a716-446655440000/settlement", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusNotFound)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "NOT_FOUND" {
		t.Errorf("code = %v; want NOT_FOUND", errObj["code"])
	}
}

func TestSettlementHandler_GetMonthly_InvalidHouseholdID(t *testing.T) {
	t.Parallel()

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	memberRepo := newMockMemberRepo()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: &mockCalculateSettlement{},
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/not-a-uuid/settlement", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	// Returns 403 Forbidden (not 400) because the authz middleware doesn't
	// reveal whether a household exists or not - it simply denies access
	// if the user is not a member.
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusForbidden)
	}
}

func TestSettlementHandler_GetMonthly_Unauthorized(t *testing.T) {
	t.Parallel()

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: &mockCalculateSettlement{},
		},
		JWTValidator:   &mockTokenValidator{returnErr: shared.ErrInvalidCredentials},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/550e8400-e29b-41d4-a716-446655440000/settlement", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSettlementHandler_GetMonthly_NotMemberOfHousehold(t *testing.T) {
	t.Parallel()

	memberRepo := newMockMemberRepo()
	// User is member of different household
	memberRepo.seedMember("m-1", "different-household-id", "user-123")

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: &mockCalculateSettlement{},
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/550e8400-e29b-41d4-a716-446655440000/settlement", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusForbidden)
	}
}

func TestSettlementHandler_GetMonthly_InvalidMonth(t *testing.T) {
	t.Parallel()

	calculateUC := &mockCalculateSettlement{
		returnErr: inbound.ErrSettlementInvalidMonth,
	}

	memberRepo := newMockMemberRepo()
	memberRepo.seedMember("m-1", "550e8400-e29b-41d4-a716-446655440000", "user-123")

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Settlement: httpadapter.SettlementHandlerDeps{
			Calculate: calculateUC,
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/550e8400-e29b-41d4-a716-446655440000/settlement?month=13", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}

	resp := parseJSONResponse(t, rec)
	errObj, _ := resp["error"].(map[string]any)
	if errObj["code"] != "VALIDATION_ERROR" {
		t.Errorf("code = %v; want VALIDATION_ERROR", errObj["code"])
	}
}
