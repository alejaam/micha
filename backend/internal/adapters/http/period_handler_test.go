package httpadapter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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

// --- Mocks for period use cases ---

type mockGetPeriodConsensus struct {
	returnOutput inbound.GetPeriodConsensusOutput
	returnErr    error
	lastInput    inbound.GetPeriodConsensusInput
}

func (m *mockGetPeriodConsensus) Execute(_ context.Context, input inbound.GetPeriodConsensusInput) (inbound.GetPeriodConsensusOutput, error) {
	m.lastInput = input
	return m.returnOutput, m.returnErr
}

type mockTransitionToReview struct{}

func (m *mockTransitionToReview) Execute(_ context.Context, _ inbound.TransitionToReviewInput) (inbound.TransitionToReviewOutput, error) {
	return inbound.TransitionToReviewOutput{}, nil
}

type mockApprovePeriod struct{}

func (m *mockApprovePeriod) Execute(_ context.Context, _ inbound.ApprovePeriodInput) (inbound.ApprovePeriodOutput, error) {
	return inbound.ApprovePeriodOutput{}, nil
}

type mockClosePeriod struct{}

func (m *mockClosePeriod) Execute(_ context.Context, _ inbound.ClosePeriodInput) (inbound.ClosePeriodOutput, error) {
	return inbound.ClosePeriodOutput{}, nil
}

type mockInitializePeriod struct{}

func (m *mockInitializePeriod) Execute(_ context.Context, _ inbound.InitializePeriodInput) (inbound.InitializePeriodOutput, error) {
	return inbound.InitializePeriodOutput{}, nil
}

// --- Tests ---

func TestPeriodHandler_GetConsensus_Success(t *testing.T) {
	t.Parallel()

	consensusUC := &mockGetPeriodConsensus{
		returnOutput: inbound.GetPeriodConsensusOutput{
			Approved: 3,
			Total:    5,
			Percent:  60.0,
		},
	}

	memberRepo := newMockMemberRepo()
	memberRepo.seedMember("m-1", "hh-1", "user-123")

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Period: httpadapter.PeriodHandlerDeps{
			TransitionToReview: &mockTransitionToReview{},
			ApprovePeriod:      &mockApprovePeriod{},
			ClosePeriod:        &mockClosePeriod{},
			InitializePeriod:   &mockInitializePeriod{},
			GetConsensus:       consensusUC,
			PeriodRepo:         newMockPeriodRepository(),
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/hh-1/periods/per-1/consensus", nil)
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

	if data["approved"].(float64) != 3 {
		t.Errorf("approved = %v; want 3", data["approved"])
	}
	if data["total"].(float64) != 5 {
		t.Errorf("total = %v; want 5", data["total"])
	}
	if data["percent"].(float64) != 60.0 {
		t.Errorf("percent = %v; want 60", data["percent"])
	}
}

func TestPeriodHandler_GetConsensus_NotFound(t *testing.T) {
	t.Parallel()

	consensusUC := &mockGetPeriodConsensus{
		returnErr: shared.ErrNotFound,
	}

	memberRepo := newMockMemberRepo()
	memberRepo.seedMember("m-1", "hh-1", "user-123")

	validator := &mockTokenValidator{
		returnUserID: "user-123",
		returnEmail:  "test@example.com",
	}

	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{
			Register: &mockRegisterUser{},
			Login:    &mockLogin{},
		},
		Period: httpadapter.PeriodHandlerDeps{
			TransitionToReview: &mockTransitionToReview{},
			ApprovePeriod:      &mockApprovePeriod{},
			ClosePeriod:        &mockClosePeriod{},
			InitializePeriod:   &mockInitializePeriod{},
			GetConsensus:       consensusUC,
			PeriodRepo:         newMockPeriodRepository(),
		},
		JWTValidator:   validator,
		MemberRepo:     memberRepo,
		AllowedOrigins: []string{"*"},
	})

	req := makeJSONRequest(t, "GET", "/v1/households/hh-1/periods/per-1/consensus", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d, body: %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
