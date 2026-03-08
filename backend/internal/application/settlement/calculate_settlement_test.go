package settlementapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	settlementapp "micha/backend/internal/application/settlement"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

func TestCalculateSettlementUseCase_Execute_Success(t *testing.T) {
	t.Parallel()
	hh := newHouseholdMock(t)
	mm := newMemberMock(t)
	ee := newExpenseMock(t)
	uc := settlementapp.NewCalculateSettlementUseCase(hh, mm, ee)

	out, err := uc.Execute(context.Background(), inbound.CalculateSettlementInput{
		HouseholdID: "hh-1",
		Year:        2026,
		Month:       3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.TotalSharedCents != 15000 {
		t.Fatalf("total shared = %d, want 15000", out.TotalSharedCents)
	}
	if len(out.Transfers) != 1 {
		t.Fatalf("transfers = %d, want 1", len(out.Transfers))
	}
}

func TestCalculateSettlementUseCase_Execute_HouseholdNotFound(t *testing.T) {
	t.Parallel()
	hh := newHouseholdMock(t)
	hh.findErr = shared.ErrNotFound
	mm := newMemberMock(t)
	ee := newExpenseMock(t)
	uc := settlementapp.NewCalculateSettlementUseCase(hh, mm, ee)

	_, err := uc.Execute(context.Background(), inbound.CalculateSettlementInput{
		HouseholdID: "missing",
		Year:        2026,
		Month:       3,
	})
	if !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCalculateSettlementUseCase_Execute_InvalidMonth(t *testing.T) {
	t.Parallel()
	hh := newHouseholdMock(t)
	mm := newMemberMock(t)
	ee := newExpenseMock(t)
	uc := settlementapp.NewCalculateSettlementUseCase(hh, mm, ee)

	_, err := uc.Execute(context.Background(), inbound.CalculateSettlementInput{
		HouseholdID: "hh-1",
		Year:        2026,
		Month:       13,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

type householdMock struct {
	house   household.Household
	findErr error
}

func newHouseholdMock(t *testing.T) *householdMock {
	t.Helper()
	h, err := household.New(household.ID("hh-1"), "Casa", household.SettlementModeEqual, "MXN", time.Now())
	if err != nil {
		t.Fatalf("household.New: %v", err)
	}
	return &householdMock{house: h}
}

func (m *householdMock) Save(context.Context, household.Household) error { return nil }
func (m *householdMock) FindByID(_ context.Context, id string) (household.Household, error) {
	if m.findErr != nil {
		return household.Household{}, m.findErr
	}
	if id != string(m.house.ID()) {
		return household.Household{}, shared.ErrNotFound
	}
	return m.house, nil
}
func (m *householdMock) List(context.Context, int, int) ([]household.Household, error) {
	return []household.Household{m.house}, nil
}
func (m *householdMock) Update(context.Context, household.Household) error { return nil }

type memberMock struct {
	members []member.Member
}

func newMemberMock(t *testing.T) *memberMock {
	t.Helper()
	now := time.Now()
	m1, err := member.New(member.ID("m-1"), "hh-1", "Ana", "ana@mail.com", 300000, now)
	if err != nil {
		t.Fatalf("member.New m1: %v", err)
	}
	m2, err := member.New(member.ID("m-2"), "hh-1", "Luis", "luis@mail.com", 100000, now)
	if err != nil {
		t.Fatalf("member.New m2: %v", err)
	}
	return &memberMock{members: []member.Member{m1, m2}}
}

func (m *memberMock) Save(context.Context, member.Member) error { return nil }
func (m *memberMock) FindByID(_ context.Context, id string) (member.Member, error) {
	for _, item := range m.members {
		if string(item.ID()) == id {
			return item, nil
		}
	}
	return member.Member{}, shared.ErrNotFound
}
func (m *memberMock) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	if householdID != "hh-1" {
		return []member.Member{}, nil
	}
	return m.members, nil
}
func (m *memberMock) ListByHousehold(context.Context, string, int, int) ([]member.Member, error) {
	return m.members, nil
}
func (m *memberMock) Update(context.Context, member.Member) error { return nil }

type expenseMock struct {
	expenses []expense.Expense
}

func newExpenseMock(t *testing.T) *expenseMock {
	t.Helper()
	now := time.Date(2026, time.March, 5, 12, 0, 0, 0, time.UTC)
	e1, err := expense.NewFromAttributes(expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-1", AmountCents: 10000, Description: "food", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodCash, CreatedAt: now})
	if err != nil {
		t.Fatalf("expense.NewFromAttributes e1: %v", err)
	}
	e2, err := expense.NewFromAttributes(expense.ExpenseAttributes{ID: "e-2", HouseholdID: "hh-1", PaidByMemberID: "m-2", AmountCents: 5000, Description: "uber", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodCard, CreatedAt: now})
	if err != nil {
		t.Fatalf("expense.NewFromAttributes e2: %v", err)
	}
	return &expenseMock{expenses: []expense.Expense{e1, e2}}
}

func (m *expenseMock) Save(context.Context, expense.Expense) error { return nil }
func (m *expenseMock) FindByID(_ context.Context, id string) (expense.Expense, error) {
	for _, item := range m.expenses {
		if string(item.ID()) == id {
			return item, nil
		}
	}
	return expense.Expense{}, shared.ErrNotFound
}
func (m *expenseMock) List(context.Context, string, int, int) ([]expense.Expense, error) {
	return m.expenses, nil
}
func (m *expenseMock) Update(context.Context, expense.Expense) error { return nil }
func (m *expenseMock) ListByHouseholdAndPeriod(_ context.Context, householdID string, from, to time.Time) ([]expense.Expense, error) {
	out := make([]expense.Expense, 0)
	for _, e := range m.expenses {
		if e.HouseholdID() == householdID && !e.CreatedAt().Before(from) && e.CreatedAt().Before(to) {
			out = append(out, e)
		}
	}
	return out, nil
}
