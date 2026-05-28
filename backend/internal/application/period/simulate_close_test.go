package periodapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// --- Mocks ---

type mockSimPeriodRepo struct {
	periods map[period.ID]period.Period
}

func newMockSimPeriodRepo() *mockSimPeriodRepo {
	return &mockSimPeriodRepo{periods: make(map[period.ID]period.Period)}
}

func (m *mockSimPeriodRepo) Create(_ context.Context, p period.Period) error {
	m.periods[p.ID()] = p
	return nil
}

func (m *mockSimPeriodRepo) GetByID(_ context.Context, id period.ID) (period.Period, error) {
	p, ok := m.periods[id]
	if !ok {
		return period.Period{}, shared.ErrNotFound
	}
	return p, nil
}

func (m *mockSimPeriodRepo) Update(_ context.Context, p period.Period) error {
	m.periods[p.ID()] = p
	return nil
}

func (m *mockSimPeriodRepo) GetCurrentOpen(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockSimPeriodRepo) GetLatestByHousehold(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockSimPeriodRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]period.Period, error) {
	return nil, nil
}

type mockSimHouseholdRepo struct {
	households map[string]household.Household
}

func newMockSimHouseholdRepo() *mockSimHouseholdRepo {
	return &mockSimHouseholdRepo{households: make(map[string]household.Household)}
}

func (m *mockSimHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	h, ok := m.households[id]
	if !ok {
		return household.Household{}, shared.ErrNotFound
	}
	return h, nil
}

func (m *mockSimHouseholdRepo) Save(_ context.Context, _ household.Household) error { return nil }
func (m *mockSimHouseholdRepo) List(_ context.Context, _, _ int) ([]household.Household, error) {
	return nil, nil
}
func (m *mockSimHouseholdRepo) ListByUserID(_ context.Context, _ string, _, _ int) ([]household.Household, error) {
	return nil, nil
}
func (m *mockSimHouseholdRepo) Update(_ context.Context, _ household.Household) error { return nil }

type mockSimMemberRepo struct {
	members []member.Member
}

func (m *mockSimMemberRepo) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID && mem.UserID() == userID {
			return mem, nil
		}
	}
	return member.Member{}, shared.ErrNotFound
}

func (m *mockSimMemberRepo) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	var result []member.Member
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *mockSimMemberRepo) Save(_ context.Context, _ member.Member) error     { return nil }
func (m *mockSimMemberRepo) FindByID(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *mockSimMemberRepo) FindByUserIDGlobal(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *mockSimMemberRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]member.Member, error) {
	return nil, nil
}
func (m *mockSimMemberRepo) ListHouseholdIDsByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockSimMemberRepo) Update(_ context.Context, _ member.Member) error  { return nil }
func (m *mockSimMemberRepo) Delete(_ context.Context, _ string) error         { return nil }
func (m *mockSimMemberRepo) CountActiveByHousehold(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *mockSimMemberRepo) LinkByEmail(_ context.Context, _, _ string) error { return nil }

type mockSimExpenseRepo struct {
	expenses []expense.Expense
}

func (m *mockSimExpenseRepo) ListByPeriod(_ context.Context, _ string) ([]expense.Expense, error) {
	return m.expenses, nil
}

func (m *mockSimExpenseRepo) FindByID(_ context.Context, id string) (expense.Expense, error) {
	for _, e := range m.expenses {
		if string(e.ID()) == id {
			return e, nil
		}
	}
	return expense.Expense{}, shared.ErrNotFound
}

func (m *mockSimExpenseRepo) Save(_ context.Context, _ expense.Expense) error { return nil }
func (m *mockSimExpenseRepo) List(_ context.Context, _ string, _, _ int) ([]expense.Expense, error) {
	return nil, nil
}
func (m *mockSimExpenseRepo) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]expense.Expense, error) {
	return nil, nil
}
func (m *mockSimExpenseRepo) SumPersonalByMemberAndPeriod(_ context.Context, _, _ string, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockSimExpenseRepo) Update(_ context.Context, _ expense.Expense) error { return nil }
func (m *mockSimExpenseRepo) AdoptOrphanExpenses(_ context.Context, _, _ string, _, _ time.Time) error {
	return nil
}

type mockSimInstallmentRepo struct {
	installments []installment.Installment
}

func (m *mockSimInstallmentRepo) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]installment.Installment, error) {
	return m.installments, nil
}

func (m *mockSimInstallmentRepo) Save(_ context.Context, _ installment.Installment) error { return nil }
func (m *mockSimInstallmentRepo) SaveAll(_ context.Context, _ []installment.Installment) error {
	return nil
}
func (m *mockSimInstallmentRepo) ListByExpense(_ context.Context, _ string) ([]installment.Installment, error) {
	return nil, nil
}
func (m *mockSimInstallmentRepo) DeleteByExpense(_ context.Context, _ string) error { return nil }

// --- Fixtures ---

func makeTestOpenPeriod(t *testing.T, id, householdID string, start, end time.Time) period.Period {
	t.Helper()
	p, err := period.NewFromAttributes(period.PeriodAttributes{
		ID:          period.ID(id),
		HouseholdID: householdID,
		StartDate:   start,
		EndDate:     end,
		Status:      period.StatusOpen,
		CreatedAt:   start,
		UpdatedAt:   start,
	})
	if err != nil {
		t.Fatalf("failed to create test period: %v", err)
	}
	return p
}

func makeSimTestHousehold(t *testing.T, id, ownerID string, frequency string, closingDay int) household.Household {
	t.Helper()
	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(id),
		Name:            "Test Household",
		OwnerID:         ownerID,
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		ClosingDay:      closingDay,
		PeriodFrequency: frequency,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}
	return h
}

// --- Tests ---

func TestSimulateClose_Success(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	ownerMemberID := "m-owner"

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	periodRepo := newMockSimPeriodRepo()
	p := makeTestOpenPeriod(t, periodID, householdID,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockSimHouseholdRepo()
	householdRepo.households[householdID] = makeSimTestHousehold(t, householdID, ownerUserID, "monthly", 15)

	memberRepo := &mockSimMemberRepo{
		members: []member.Member{
			makeTestMember(t, ownerMemberID, householdID, ownerUserID),
		},
	}

	expenseRepo := &mockSimExpenseRepo{
		expenses: []expense.Expense{
			makeTestFixedExpense(t, "e-1", householdID, periodID, ownerMemberID, 50000),
		},
	}

	installmentRepo := &mockSimInstallmentRepo{
		installments: []installment.Installment{
			makeTestInstallment(t, "inst-1", "e-2", ownerMemberID, time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC)),
		},
	}

	uc := NewSimulateClosePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		expenseRepo,
		installmentRepo,
	)
	uc.now = func() time.Time { return now }

	output, err := uc.Execute(context.Background(), inbound.SimulateClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: ownerUserID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify next period dates
	if output.NextPeriodStart.IsZero() {
		t.Error("expected non-zero NextPeriodStart")
	}
	if output.NextPeriodEnd.IsZero() {
		t.Error("expected non-zero NextPeriodEnd")
	}

	// Monthly: nextStart = Feb 1, nextEnd = Mar 15 (closingDay=15, next month)
	expectedStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 3, 15, 23, 59, 59, 999999999, time.UTC)
	if !output.NextPeriodStart.Equal(expectedStart) {
		t.Errorf("NextPeriodStart = %v; want %v", output.NextPeriodStart, expectedStart)
	}
	if !output.NextPeriodEnd.Equal(expectedEnd) {
		t.Errorf("NextPeriodEnd = %v; want %v", output.NextPeriodEnd, expectedEnd)
	}

	// Verify rollover counts
	if output.FixedExpenseCount != 1 {
		t.Errorf("FixedExpenseCount = %d; want 1", output.FixedExpenseCount)
	}
	if output.InstallmentCount != 1 {
		t.Errorf("InstallmentCount = %d; want 1", output.InstallmentCount)
	}
}

func TestSimulateClose_NotOwnerGetsForbidden(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	nonOwnerUserID := "user-non-owner"

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	periodRepo := newMockSimPeriodRepo()
	p := makeTestOpenPeriod(t, periodID, householdID,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockSimHouseholdRepo()
	householdRepo.households[householdID] = makeSimTestHousehold(t, householdID, ownerUserID, "monthly", 15)

	memberRepo := &mockSimMemberRepo{}

	uc := NewSimulateClosePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		&mockSimExpenseRepo{},
		&mockSimInstallmentRepo{},
	)
	uc.now = func() time.Time { return now }

	_, err := uc.Execute(context.Background(), inbound.SimulateClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: nonOwnerUserID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, shared.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestSimulateClose_FuturePeriod(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	ownerMemberID := "m-owner"

	// Period ends Jan 31, so nextStart = Feb 1.
	// If today is Jan 31, Feb 1 is in the future.
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	periodRepo := newMockSimPeriodRepo()
	p := makeTestOpenPeriod(t, periodID, householdID,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockSimHouseholdRepo()
	householdRepo.households[householdID] = makeSimTestHousehold(t, householdID, ownerUserID, "monthly", 15)

	memberRepo := &mockSimMemberRepo{
		members: []member.Member{
			makeTestMember(t, ownerMemberID, householdID, ownerUserID),
		},
	}

	uc := NewSimulateClosePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		&mockSimExpenseRepo{},
		&mockSimInstallmentRepo{},
	)
	uc.now = func() time.Time { return now }

	_, err := uc.Execute(context.Background(), inbound.SimulateClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: ownerUserID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, shared.ErrFuturePeriod) {
		t.Errorf("expected ErrFuturePeriod, got: %v", err)
	}
}

func TestSimulateClose_PeriodTooShort(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	ownerMemberID := "m-owner"

	// Period started 3 days ago.
	now := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)

	periodRepo := newMockSimPeriodRepo()
	p := makeTestOpenPeriod(t, periodID, householdID,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockSimHouseholdRepo()
	householdRepo.households[householdID] = makeSimTestHousehold(t, householdID, ownerUserID, "monthly", 15)

	memberRepo := &mockSimMemberRepo{
		members: []member.Member{
			makeTestMember(t, ownerMemberID, householdID, ownerUserID),
		},
	}

	uc := NewSimulateClosePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		&mockSimExpenseRepo{},
		&mockSimInstallmentRepo{},
	)
	uc.now = func() time.Time { return now }

	_, err := uc.Execute(context.Background(), inbound.SimulateClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: ownerUserID,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, period.ErrPeriodTooShort) {
		t.Errorf("expected ErrPeriodTooShort, got: %v", err)
	}
}

// Helper to create test fixed expenses
func makeTestFixedExpense(t *testing.T, id, householdID, periodID, memberID string, amountCents int64) expense.Expense {
	t.Helper()
	e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:                expense.ID(id),
		HouseholdID:       householdID,
		PaidByMemberID:    memberID,
		PeriodID:          periodID,
		CategoryID:        "cat-1",
		AmountCents:       amountCents,
		Description:       "Test fixed expense",
		IsShared:          true,
		Currency:          "MXN",
		PaymentMethod:     expense.PaymentMethodCash,
		ExpenseType:       expense.ExpenseTypeFixed,
		TotalInstallments: 0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test expense: %v", err)
	}
	return e
}

// Helper to create test installments
func makeTestInstallment(t *testing.T, id, expenseID, memberID string, startDate time.Time) installment.Installment {
	t.Helper()
	inst, err := installment.NewFromAttributes(installment.InstallmentAttributes{
		ID:                     installment.ID(id),
		ExpenseID:              expenseID,
		PaidByMemberID:         memberID,
		TotalInstallments:      3,
		CurrentInstallment:     2,
		InstallmentAmountCents: 10000,
		TotalAmountCents:       30000,
		StartDate:              startDate,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test installment: %v", err)
	}
	return inst
}
