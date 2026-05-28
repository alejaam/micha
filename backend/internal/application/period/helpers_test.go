package periodapp

import (
	"context"
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
)

// ---------------------------------------------------------------------------
// Shared Mocks (used by initialize_period_test.go and simulate_close_test.go)
// ---------------------------------------------------------------------------

type mockPeriodRepo struct {
	periods map[period.ID]period.Period
}

func newMockPeriodRepo() *mockPeriodRepo {
	return &mockPeriodRepo{periods: make(map[period.ID]period.Period)}
}

func (m *mockPeriodRepo) Create(_ context.Context, p period.Period) error {
	m.periods[p.ID()] = p
	return nil
}

func (m *mockPeriodRepo) GetByID(_ context.Context, id period.ID) (period.Period, error) {
	p, ok := m.periods[id]
	if !ok {
		return period.Period{}, shared.ErrNotFound
	}
	return p, nil
}

func (m *mockPeriodRepo) Update(_ context.Context, p period.Period) error {
	m.periods[p.ID()] = p
	return nil
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

type mockHouseholdRepo struct {
	households map[string]household.Household
}

func newMockHouseholdRepo() *mockHouseholdRepo {
	return &mockHouseholdRepo{households: make(map[string]household.Household)}
}

func (m *mockHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	h, ok := m.households[id]
	if !ok {
		return household.Household{}, shared.ErrNotFound
	}
	return h, nil
}

func (m *mockHouseholdRepo) Save(_ context.Context, _ household.Household) error { return nil }
func (m *mockHouseholdRepo) List(_ context.Context, _, _ int) ([]household.Household, error) {
	return nil, nil
}
func (m *mockHouseholdRepo) ListByUserID(_ context.Context, _ string, _, _ int) ([]household.Household, error) {
	return nil, nil
}
func (m *mockHouseholdRepo) Update(_ context.Context, _ household.Household) error { return nil }

type mockMemberRepo2 struct {
	members []member.Member
}

func (m *mockMemberRepo2) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID && mem.UserID() == userID {
			return mem, nil
		}
	}
	return member.Member{}, shared.ErrNotFound
}

func (m *mockMemberRepo2) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	var result []member.Member
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *mockMemberRepo2) Save(_ context.Context, _ member.Member) error     { return nil }
func (m *mockMemberRepo2) FindByID(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *mockMemberRepo2) FindByUserIDGlobal(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *mockMemberRepo2) ListByHousehold(_ context.Context, _ string, _, _ int) ([]member.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo2) ListHouseholdIDsByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockMemberRepo2) Update(_ context.Context, _ member.Member) error  { return nil }
func (m *mockMemberRepo2) Delete(_ context.Context, _ string) error         { return nil }
func (m *mockMemberRepo2) CountActiveByHousehold(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *mockMemberRepo2) LinkByEmail(_ context.Context, _, _ string) error { return nil }

type mockExpenseRepo struct{}

func (m *mockExpenseRepo) ListByPeriod(_ context.Context, _ string) ([]expense.Expense, error) {
	return nil, nil
}
func (m *mockExpenseRepo) FindByID(_ context.Context, _ string) (expense.Expense, error) {
	return expense.Expense{}, shared.ErrNotFound
}
func (m *mockExpenseRepo) Save(_ context.Context, _ expense.Expense) error { return nil }
func (m *mockExpenseRepo) List(_ context.Context, _ string, _, _ int) ([]expense.Expense, error) {
	return nil, nil
}
func (m *mockExpenseRepo) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]expense.Expense, error) {
	return nil, nil
}
func (m *mockExpenseRepo) SumPersonalByMemberAndPeriod(_ context.Context, _, _ string, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockExpenseRepo) Update(_ context.Context, _ expense.Expense) error { return nil }
func (m *mockExpenseRepo) AdoptOrphanExpenses(_ context.Context, _, _ string, _, _ time.Time) error {
	return nil
}

type mockInstallmentRepo struct{}

func (m *mockInstallmentRepo) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]installment.Installment, error) {
	return nil, nil
}
func (m *mockInstallmentRepo) Save(_ context.Context, _ installment.Installment) error { return nil }
func (m *mockInstallmentRepo) SaveAll(_ context.Context, _ []installment.Installment) error {
	return nil
}
func (m *mockInstallmentRepo) ListByExpense(_ context.Context, _ string) ([]installment.Installment, error) {
	return nil, nil
}
func (m *mockInstallmentRepo) DeleteByExpense(_ context.Context, _ string) error { return nil }

type mockTxManager struct{}

func (m *mockTxManager) Run(_ context.Context, fn func(ctx context.Context) error) error {
	return fn(context.Background())
}

type mockIDGen struct{ counter int }

func (m *mockIDGen) NewID() string {
	m.counter++
	return "id-" + string(rune('a'+m.counter-1))
}

// --- Fixture helpers ---

func makeTestMember(t *testing.T, id, householdID, userID string) member.Member {
	t.Helper()
	m, err := member.NewFromAttributes(member.Attributes{
		ID:          member.ID(id),
		HouseholdID: householdID,
		Name:        "Test Member",
		Email:       "test@example.com",
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test member: %v", err)
	}
	return m
}
