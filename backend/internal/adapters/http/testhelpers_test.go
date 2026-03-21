package httpadapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// --- Test Helpers ---

func makeJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode request body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func parseJSONResponse(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response JSON: %v, body: %s", err, rec.Body.String())
	}
	return result
}

// --- Mock Use Cases ---

// mockRegisterUser implements inbound.RegisterUserUseCase
type mockRegisterUser struct {
	returnOutput inbound.RegisterUserOutput
	returnErr    error
	lastInput    inbound.RegisterUserInput
}

func (m *mockRegisterUser) Execute(_ context.Context, input inbound.RegisterUserInput) (inbound.RegisterUserOutput, error) {
	m.lastInput = input
	return m.returnOutput, m.returnErr
}

// mockLogin implements inbound.LoginUseCase
type mockLogin struct {
	returnOutput inbound.LoginOutput
	returnErr    error
	lastInput    inbound.LoginInput
}

func (m *mockLogin) Execute(_ context.Context, input inbound.LoginInput) (inbound.LoginOutput, error) {
	m.lastInput = input
	return m.returnOutput, m.returnErr
}

// mockRegisterExpense implements inbound.RegisterExpenseUseCase
type mockRegisterExpense struct {
	returnOutput inbound.RegisterExpenseOutput
	returnErr    error
	lastInput    inbound.RegisterExpenseInput
}

func (m *mockRegisterExpense) Execute(_ context.Context, input inbound.RegisterExpenseInput) (inbound.RegisterExpenseOutput, error) {
	m.lastInput = input
	return m.returnOutput, m.returnErr
}

// mockGetExpense implements inbound.GetExpenseUseCase
type mockGetExpense struct {
	returnExpense expense.Expense
	returnErr     error
}

func (m *mockGetExpense) Execute(_ context.Context, _ string) (expense.Expense, error) {
	return m.returnExpense, m.returnErr
}

// mockListExpenses implements inbound.ListExpensesUseCase
type mockListExpenses struct {
	returnExpenses []expense.Expense
	returnErr      error
}

func (m *mockListExpenses) Execute(_ context.Context, _ inbound.ListExpensesQuery) ([]expense.Expense, error) {
	return m.returnExpenses, m.returnErr
}

// mockPatchExpense implements inbound.PatchExpenseUseCase
type mockPatchExpense struct {
	returnExpense expense.Expense
	returnErr     error
	lastInput     inbound.PatchExpenseCommand
}

func (m *mockPatchExpense) Execute(_ context.Context, input inbound.PatchExpenseCommand) (expense.Expense, error) {
	m.lastInput = input
	return m.returnExpense, m.returnErr
}

// mockDeleteExpense implements inbound.DeleteExpenseUseCase
type mockDeleteExpense struct {
	returnErr error
}

func (m *mockDeleteExpense) Execute(_ context.Context, _ string) error {
	return m.returnErr
}

// mockCalculateSettlement implements inbound.CalculateSettlementUseCase
type mockCalculateSettlement struct {
	returnOutput inbound.CalculateSettlementOutput
	returnErr    error
	lastInput    inbound.CalculateSettlementInput
}

func (m *mockCalculateSettlement) Execute(_ context.Context, input inbound.CalculateSettlementInput) (inbound.CalculateSettlementOutput, error) {
	m.lastInput = input
	return m.returnOutput, m.returnErr
}

// mockRegisterHousehold implements inbound.RegisterHouseholdUseCase
type mockRegisterHousehold struct {
	returnOutput inbound.RegisterHouseholdOutput
	returnErr    error
}

func (m *mockRegisterHousehold) Execute(_ context.Context, _ inbound.RegisterHouseholdInput) (inbound.RegisterHouseholdOutput, error) {
	return m.returnOutput, m.returnErr
}

// mockListHouseholds implements inbound.ListHouseholdsUseCase
type mockListHouseholds struct {
	returnHouseholds []household.Household
	returnErr        error
}

func (m *mockListHouseholds) Execute(_ context.Context, _ inbound.ListHouseholdsQuery) ([]household.Household, error) {
	return m.returnHouseholds, m.returnErr
}

// mockGetHousehold implements inbound.GetHouseholdUseCase
type mockGetHousehold struct {
	returnHousehold household.Household
	returnErr       error
}

func (m *mockGetHousehold) Execute(_ context.Context, _ string) (household.Household, error) {
	return m.returnHousehold, m.returnErr
}

// mockUpdateHousehold implements inbound.UpdateHouseholdUseCase
type mockUpdateHousehold struct {
	returnErr error
}

func (m *mockUpdateHousehold) Execute(_ context.Context, _ inbound.UpdateHouseholdInput) error {
	return m.returnErr
}

// mockRegisterMember implements inbound.RegisterMemberUseCase
type mockRegisterMember struct {
	returnOutput inbound.RegisterMemberOutput
	returnErr    error
}

func (m *mockRegisterMember) Execute(_ context.Context, _ inbound.RegisterMemberInput) (inbound.RegisterMemberOutput, error) {
	return m.returnOutput, m.returnErr
}

// mockListMembers implements inbound.ListMembersUseCase
type mockListMembers struct {
	returnMembers []member.Member
	returnErr     error
}

func (m *mockListMembers) Execute(_ context.Context, _ inbound.ListMembersQuery) ([]member.Member, error) {
	return m.returnMembers, m.returnErr
}

// mockUpdateMember implements inbound.UpdateMemberUseCase
type mockUpdateMember struct {
	returnErr error
}

func (m *mockUpdateMember) Execute(_ context.Context, _ inbound.UpdateMemberInput) error {
	return m.returnErr
}

// mockDeleteMember implements inbound.DeleteMemberUseCase
type mockDeleteMember struct {
	returnErr error
}

func (m *mockDeleteMember) Execute(_ context.Context, _ inbound.DeleteMemberInput) error {
	return m.returnErr
}

// mockUpdateSplitConfig implements inbound.UpdateSplitConfigUseCase
type mockUpdateSplitConfig struct {
	returnErr error
}

func (m *mockUpdateSplitConfig) Execute(_ context.Context, _ inbound.UpdateSplitConfigInput) error {
	return m.returnErr
}

// mockCreateCategory implements inbound.CreateCategoryUseCase
type mockCreateCategory struct {
	returnOutput inbound.CreateCategoryOutput
	returnErr    error
}

func (m *mockCreateCategory) Execute(_ context.Context, _ inbound.CreateCategoryInput) (inbound.CreateCategoryOutput, error) {
	return m.returnOutput, m.returnErr
}

// mockListCategories implements inbound.ListCategoriesUseCase
type mockListCategories struct {
	returnCategories []category.Category
	returnErr        error
}

func (m *mockListCategories) Execute(_ context.Context, _ inbound.ListCategoriesQuery) ([]category.Category, error) {
	return m.returnCategories, m.returnErr
}

// mockDeleteCategory implements inbound.DeleteCategoryUseCase
type mockDeleteCategory struct {
	returnErr error
}

func (m *mockDeleteCategory) Execute(_ context.Context, _ inbound.DeleteCategoryInput) error {
	return m.returnErr
}

// --- Mock Token Validator ---

type mockTokenValidator struct {
	returnUserID string
	returnEmail  string
	returnErr    error
}

func (m *mockTokenValidator) Validate(_ string) (string, string, error) {
	return m.returnUserID, m.returnEmail, m.returnErr
}

// --- Mock Member Repository (for authz middleware) ---

type mockMemberRepo struct {
	members map[string]member.Member
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[string]member.Member)}
}

func (m *mockMemberRepo) seedMember(id, householdID, userID string) {
	mem, _ := member.NewFromAttributes(member.Attributes{
		ID:          member.ID(id),
		HouseholdID: householdID,
		Name:        "Test",
		Email:       "test@example.com",
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	m.members[id] = mem
}

func (m *mockMemberRepo) Save(_ context.Context, mem member.Member) error {
	m.members[string(mem.ID())] = mem
	return nil
}

func (m *mockMemberRepo) FindByID(_ context.Context, id string) (member.Member, error) {
	if mem, ok := m.members[id]; ok {
		return mem, nil
	}
	return member.Member{}, shared.ErrNotFound
}

func (m *mockMemberRepo) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID && mem.UserID() == userID {
			return mem, nil
		}
	}
	return member.Member{}, shared.ErrNotFound
}

func (m *mockMemberRepo) FindByUserIDGlobal(_ context.Context, userID string) (member.Member, error) {
	for _, mem := range m.members {
		if mem.UserID() == userID {
			return mem, nil
		}
	}
	return member.Member{}, shared.ErrNotFound
}

func (m *mockMemberRepo) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	var result []member.Member
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *mockMemberRepo) ListByHousehold(_ context.Context, householdID string, _, _ int) ([]member.Member, error) {
	return m.ListAllByHousehold(context.Background(), householdID)
}

func (m *mockMemberRepo) ListHouseholdIDsByUserID(_ context.Context, userID string) ([]string, error) {
	var ids []string
	for _, mem := range m.members {
		if mem.UserID() == userID {
			ids = append(ids, mem.HouseholdID())
		}
	}
	return ids, nil
}

func (m *mockMemberRepo) Update(_ context.Context, mem member.Member) error {
	m.members[string(mem.ID())] = mem
	return nil
}

func (m *mockMemberRepo) Delete(_ context.Context, id string) error {
	delete(m.members, id)
	return nil
}

func (m *mockMemberRepo) CountActiveByHousehold(_ context.Context, householdID string) (int, error) {
	count := 0
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID {
			count++
		}
	}
	return count, nil
}

// --- Helper to create test expenses ---

func makeTestExpense(t *testing.T, id, householdID string, amountCents int64) expense.Expense {
	t.Helper()
	e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID(id),
		HouseholdID:    householdID,
		PaidByMemberID: "m-1",
		AmountCents:    amountCents,
		Description:    "Test expense",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  expense.PaymentMethodCash,
		ExpenseType:    expense.ExpenseTypeVariable,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test expense: %v", err)
	}
	return e
}

// --- Helper to create test settlements ---

func makeTestSettlementOutput() inbound.CalculateSettlementOutput {
	return inbound.CalculateSettlementOutput{
		HouseholdID:          "hh-1",
		Year:                 2026,
		Month:                3,
		SettlementMode:       household.SettlementModeEqual,
		TotalSharedCents:     15000,
		IncludedExpenseCount: 2,
		Members: []inbound.MemberSettlement{
			{MemberID: "m-1", Name: "Ana", PaidCents: 10000, ExpectedShare: 7500, NetBalanceCents: 2500},
			{MemberID: "m-2", Name: "Luis", PaidCents: 5000, ExpectedShare: 7500, NetBalanceCents: -2500},
		},
		Transfers: []inbound.SettlementTransfer{
			{FromMemberID: "m-2", ToMemberID: "m-1", AmountCents: 2500},
		},
	}
}

// --- Helper to create test household ---

func makeTestHousehold(t *testing.T, id string) household.Household {
	t.Helper()
	h, err := household.NewFromAttributes(household.Attributes{
		ID:             household.ID(id),
		Name:           "Test Household",
		SettlementMode: household.SettlementModeEqual,
		Currency:       "MXN",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}
	return h
}

// --- Helper to create test member ---

func makeTestMember(t *testing.T, id, householdID string) member.Member {
	t.Helper()
	m, err := member.NewFromAttributes(member.Attributes{
		ID:                 member.ID(id),
		HouseholdID:        householdID,
		Name:               "Test Member",
		Email:              "test@example.com",
		MonthlySalaryCents: 100000,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test member: %v", err)
	}
	return m
}
