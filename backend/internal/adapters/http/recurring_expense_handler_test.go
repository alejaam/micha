package httpadapter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "micha/backend/internal/adapters/http"
	recurringexpenseapp "micha/backend/internal/application/recurringexpense"
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type mockCreateRecurringExpense struct {
	lastInput inbound.CreateRecurringExpenseInput
}

func (m *mockCreateRecurringExpense) Execute(_ context.Context, input inbound.CreateRecurringExpenseInput) (inbound.CreateRecurringExpenseOutput, error) {
	m.lastInput = input
	return inbound.CreateRecurringExpenseOutput{RecurringExpenseID: "recurring-1"}, nil
}

type mockGetRecurringExpense struct {
	lastInput inbound.GetRecurringExpenseInput
	returnErr error
}

func (m *mockGetRecurringExpense) Execute(_ context.Context, input inbound.GetRecurringExpenseInput) (recurringexpense.RecurringExpense, error) {
	m.lastInput = input
	return recurringexpense.RecurringExpense{}, m.returnErr
}

type mockListRecurringExpenses struct {
	lastInput inbound.ListRecurringExpensesQuery
}

func (m *mockListRecurringExpenses) Execute(_ context.Context, input inbound.ListRecurringExpensesQuery) ([]recurringexpense.RecurringExpense, error) {
	m.lastInput = input
	return nil, nil
}

type mockUpdateRecurringExpense struct {
	lastInput inbound.UpdateRecurringExpenseCommand
}

func (m *mockUpdateRecurringExpense) Execute(_ context.Context, input inbound.UpdateRecurringExpenseCommand) (recurringexpense.RecurringExpense, error) {
	m.lastInput = input
	return recurringexpense.RecurringExpense{}, nil
}

type mockDeleteRecurringExpense struct {
	lastInput inbound.DeleteRecurringExpenseInput
}

func (m *mockDeleteRecurringExpense) Execute(_ context.Context, input inbound.DeleteRecurringExpenseInput) error {
	m.lastInput = input
	return nil
}

type mockGenerateRecurringExpenses struct {
	lastInput inbound.GenerateRecurringExpensesCommand
}

func (m *mockGenerateRecurringExpenses) Execute(_ context.Context, input inbound.GenerateRecurringExpensesCommand) (inbound.GenerateRecurringExpensesOutput, error) {
	m.lastInput = input
	return inbound.GenerateRecurringExpensesOutput{}, nil
}

func TestRecurringExpenseHandler_PropagatesCurrentUserID(t *testing.T) {
	create := &mockCreateRecurringExpense{}
	get := &mockGetRecurringExpense{}
	list := &mockListRecurringExpenses{}
	update := &mockUpdateRecurringExpense{}
	deleteRecurringExpense := &mockDeleteRecurringExpense{}
	generate := &mockGenerateRecurringExpenses{}
	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{Register: &mockRegisterUser{}, Login: &mockLogin{}},
		RecurringExpense: httpadapter.RecurringExpenseHandlerDeps{
			Create: create,
			Get: get,
			List: list,
			Update: update,
			Delete: deleteRecurringExpense,
			Generate: generate,
		},
		JWTValidator:   &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo:     newMockMemberRepo(),
		AllowedOrigins: []string{"*"},
	})

	requests := []struct {
		method string
		path   string
		body   any
		status int
	}{
		{method: http.MethodPost, path: "/v1/recurring-expenses", body: map[string]any{"household_id": "household-1", "is_agnostic": true, "amount_cents": 1000, "description": "Rent", "expense_type": "fixed", "recurrence_pattern": "monthly", "start_date": "2026-01-01"}, status: http.StatusCreated},
		{method: http.MethodGet, path: "/v1/recurring-expenses/550e8400-e29b-41d4-a716-446655440000", status: http.StatusOK},
		{method: http.MethodGet, path: "/v1/recurring-expenses?household_id=household-1", status: http.StatusOK},
		{method: http.MethodPatch, path: "/v1/recurring-expenses/550e8400-e29b-41d4-a716-446655440000", body: map[string]any{"description": "Updated"}, status: http.StatusOK},
		{method: http.MethodDelete, path: "/v1/recurring-expenses/550e8400-e29b-41d4-a716-446655440000", status: http.StatusNoContent},
		{method: http.MethodPost, path: "/v1/recurring-expenses/generate", body: map[string]any{"household_id": "household-1", "as_of_date": "2026-01-01"}, status: http.StatusOK},
	}
	for _, tt := range requests {
		req := makeJSONRequest(t, tt.method, tt.path, tt.body)
		req.Header.Set("Authorization", "Bearer valid-token")
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		if rec.Code != tt.status {
			t.Errorf("%s %s status = %d; want %d", tt.method, tt.path, rec.Code, tt.status)
		}
	}

	for _, got := range []string{
		create.lastInput.CurrentUserID,
		get.lastInput.CurrentUserID,
		list.lastInput.CurrentUserID,
		update.lastInput.CurrentUserID,
		deleteRecurringExpense.lastInput.CurrentUserID,
		generate.lastInput.CurrentUserID,
	} {
		if got != "user-1" {
			t.Errorf("CurrentUserID = %q; want user-1", got)
		}
	}
}

type httpGenerateRecurringExpenseRepo struct {
	outbound.RecurringExpenseRepository
	listDueForGenerationCalls int
}

func (r *httpGenerateRecurringExpenseRepo) ListDueForGeneration(_ context.Context, _ time.Time) ([]recurringexpense.RecurringExpense, error) {
	r.listDueForGenerationCalls++
	return nil, nil
}

type httpGenerateIDGenerator struct{}

func (httpGenerateIDGenerator) NewID() string { return "expense-1" }

func TestRecurringExpenseHandler_GenerateRejectsBlankHouseholdID(t *testing.T) {
	for _, householdID := range []string{"", " \t "} {
		t.Run("household_id="+householdID, func(t *testing.T) {
			repo := &httpGenerateRecurringExpenseRepo{}
			server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
				Auth: httpadapter.AuthHandlerDeps{Register: &mockRegisterUser{}, Login: &mockLogin{}},
				RecurringExpense: httpadapter.RecurringExpenseHandlerDeps{
					Create: &mockCreateRecurringExpense{}, Get: &mockGetRecurringExpense{}, List: &mockListRecurringExpenses{}, Update: &mockUpdateRecurringExpense{}, Delete: &mockDeleteRecurringExpense{},
					Generate: recurringexpenseapp.NewGenerateRecurringExpensesUseCase(repo, newMockMemberRepo(), nil, httpGenerateIDGenerator{}),
				},
				JWTValidator: &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
				MemberRepo: newMockMemberRepo(),
			})

			req := makeJSONRequest(t, http.MethodPost, "/v1/recurring-expenses/generate", map[string]any{"household_id": householdID})
			req.Header.Set("Authorization", "Bearer valid-token")
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
			}
			if repo.listDueForGenerationCalls != 0 {
				t.Errorf("ListDueForGeneration calls = %d; want 0", repo.listDueForGenerationCalls)
			}
		})
	}
}

func TestRecurringExpenseHandler_MapsForbidden(t *testing.T) {
	get := &mockGetRecurringExpense{returnErr: shared.ErrForbidden}
	server := httpadapter.NewServer("8080", httpadapter.ServerDependencies{
		Auth: httpadapter.AuthHandlerDeps{Register: &mockRegisterUser{}, Login: &mockLogin{}},
		RecurringExpense: httpadapter.RecurringExpenseHandlerDeps{
			Create: &mockCreateRecurringExpense{}, Get: get, List: &mockListRecurringExpenses{}, Update: &mockUpdateRecurringExpense{}, Delete: &mockDeleteRecurringExpense{}, Generate: &mockGenerateRecurringExpenses{},
		},
		JWTValidator: &mockTokenValidator{returnUserID: "user-1", returnEmail: "test@example.com"},
		MemberRepo: newMockMemberRepo(),
	})

	req := makeJSONRequest(t, http.MethodGet, "/v1/recurring-expenses/550e8400-e29b-41d4-a716-446655440000", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusForbidden)
	}
}
