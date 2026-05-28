package httpadapter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "micha/backend/internal/adapters/http"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ensure our mock satisfies the interface at compile time
var _ outbound.PeriodRepository = (*mockPeriodRepo)(nil)

// --- Mocks ---

// mockPeriodRepo satisfies outbound.PeriodRepository, returning not found for everything.
type mockPeriodRepo struct{}

func newMockPeriodRepository() *mockPeriodRepo {
	return &mockPeriodRepo{}
}

func (m *mockPeriodRepo) Create(_ context.Context, _ period.Period) error {
	return errors.New("unexpected call to Create")
}

func (m *mockPeriodRepo) Update(_ context.Context, _ period.Period) error {
	return errors.New("unexpected call to Update")
}

func (m *mockPeriodRepo) GetByID(_ context.Context, _ period.ID) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockPeriodRepo) GetCurrentOpen(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockPeriodRepo) GetLatestByHousehold(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockPeriodRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]period.Period, error) {
	return nil, nil
}

// --- Mock for SimulateClose ---

type mockSimulateClosePeriod struct {
	returnOutput inbound.SimulateClosePeriodOutput
	returnErr    error
}

func (m *mockSimulateClosePeriod) Execute(_ context.Context, _ inbound.SimulateClosePeriodInput) (inbound.SimulateClosePeriodOutput, error) {
	return m.returnOutput, m.returnErr
}

// --- Mock for InitializePeriod ---

type mockInitializePeriod struct {
	returnOutput inbound.InitializePeriodOutput
	returnErr    error
}

func (m *mockInitializePeriod) Execute(_ context.Context, _ inbound.InitializePeriodInput) (inbound.InitializePeriodOutput, error) {
	return m.returnOutput, m.returnErr
}

// --- Test helpers ---

func newPeriodHandlerDeps(
	simulateClose inbound.SimulateClosePeriodUseCase,
	initPeriod inbound.InitializePeriodUseCase,
	periodRepo outbound.PeriodRepository,
) httpadapter.PeriodHandlerDeps {
	return httpadapter.PeriodHandlerDeps{
		SimulateClose:    simulateClose,
		InitializePeriod: initPeriod,
		PeriodRepo:       periodRepo,
	}
}

func makeTestServerWithPeriodDeps(t *testing.T, periodDeps httpadapter.PeriodHandlerDeps) httpadapter.Server {
	t.Helper()
	memberRepo := newMockMemberRepo()
	memberRepo.seedMember("m-1", "hh-1", "user-123")
	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}
	return httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Period:         periodDeps,
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})
}

// --- Tests ---

func TestPeriodHandler_SimulateClose_Success(t *testing.T) {
	t.Parallel()

	nextStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	nextEnd := time.Date(2026, 3, 15, 23, 59, 59, 999999999, time.UTC)

	simulateUC := &mockSimulateClosePeriod{
		returnOutput: inbound.SimulateClosePeriodOutput{
			NextPeriodStart:   nextStart,
			NextPeriodEnd:     nextEnd,
			SettlementPreview: []inbound.SettlementEntry{},
			FixedExpenseCount: 2,
			InstallmentCount:  1,
		},
	}

	deps := newPeriodHandlerDeps(simulateUC, &mockInitializePeriod{}, newMockPeriodRepository())
	server := makeTestServerWithPeriodDeps(t, deps)

	req := makeJSONRequest(t, "POST", "/v1/households/hh-1/periods/per-1/simulate-close", nil)
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

	if data["fixed_expense_count"].(float64) != 2 {
		t.Errorf("fixed_expense_count = %v; want 2", data["fixed_expense_count"])
	}
	if data["installment_count"].(float64) != 1 {
		t.Errorf("installment_count = %v; want 1", data["installment_count"])
	}
}

func TestPeriodHandler_SimulateClose_Forbidden(t *testing.T) {
	t.Parallel()

	simulateUC := &mockSimulateClosePeriod{
		returnErr: shared.ErrForbidden,
	}

	deps := newPeriodHandlerDeps(simulateUC, &mockInitializePeriod{}, newMockPeriodRepository())
	server := makeTestServerWithPeriodDeps(t, deps)

	req := makeJSONRequest(t, "POST", "/v1/households/hh-1/periods/per-1/simulate-close", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want %d, body: %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPeriodHandler_SimulateClose_FuturePeriod(t *testing.T) {
	t.Parallel()

	simulateUC := &mockSimulateClosePeriod{
		returnErr: shared.ErrFuturePeriod,
	}

	deps := newPeriodHandlerDeps(simulateUC, &mockInitializePeriod{}, newMockPeriodRepository())
	server := makeTestServerWithPeriodDeps(t, deps)

	req := makeJSONRequest(t, "POST", "/v1/households/hh-1/periods/per-1/simulate-close", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d, body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	resp := parseJSONResponse(t, rec)
	if code, ok := resp["error"].(map[string]any)["code"]; ok {
		if code != "FUTURE_PERIOD" {
			t.Errorf("error code = %v; want FUTURE_PERIOD", code)
		}
	}
}
