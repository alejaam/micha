package settlementapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	settlementapp "micha/backend/internal/application/settlement"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

func TestCalculateSettlementUseCase_Execute_Success(t *testing.T) {
	t.Parallel()
	hh := newHouseholdMock(t)
	mm := newMemberMock(t)
	ee := newExpenseMock(t)
	ii := &installmentMock{}
	rr := &recurringExpenseMock{}
	uc := settlementapp.NewCalculateSettlementUseCase(hh, mm, ee, ii, rr)

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
	ii := &installmentMock{}
	rr := &recurringExpenseMock{}
	uc := settlementapp.NewCalculateSettlementUseCase(hh, mm, ee, ii, rr)

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
	ii := &installmentMock{}
	rr := &recurringExpenseMock{}
	uc := settlementapp.NewCalculateSettlementUseCase(hh, mm, ee, ii, rr)

	_, err := uc.Execute(context.Background(), inbound.CalculateSettlementInput{
		HouseholdID: "hh-1",
		Year:        2026,
		Month:       13,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCalculateSettlementUseCase_Execute_IgnoresAgnosticFixedRecurringWithoutPayment(t *testing.T) {
	t.Parallel()
	hh := newHouseholdMock(t)
	mm := newMemberMock(t)
	ee := newExpenseMock(t)
	ii := &installmentMock{}

	start := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	re, err := recurringexpense.NewFromAttributes(recurringexpense.RecurringExpenseAttributes{
		ID:                 "re-1",
		HouseholdID:        "hh-1",
		PaidByMemberID:     "",
		IsAgnostic:         true,
		AmountCents:        10000,
		Description:        "Rent",
		CategoryID:         "cat-rent",
		ExpenseType:        expense.ExpenseTypeFixed,
		RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
		StartDate:          start,
		NextGenerationDate: start,
		IsActive:           true,
		CreatedAt:          start,
		UpdatedAt:          start,
	})
	if err != nil {
		t.Fatalf("recurringexpense.NewFromAttributes: %v", err)
	}

	rr := &recurringExpenseMock{items: []recurringexpense.RecurringExpense{re}}
	uc := settlementapp.NewCalculateSettlementUseCase(hh, mm, ee, ii, rr)

	out, err := uc.Execute(context.Background(), inbound.CalculateSettlementInput{
		HouseholdID: "hh-1",
		Year:        2026,
		Month:       3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Settlement only includes paid expenses/installments.
	if out.TotalSharedCents != 15000 {
		t.Fatalf("total shared = %d, want 15000", out.TotalSharedCents)
	}
}

type installmentMock struct{}

func (m *installmentMock) Save(context.Context, installment.Installment) error      { return nil }
func (m *installmentMock) SaveAll(context.Context, []installment.Installment) error { return nil }
func (m *installmentMock) ListByExpense(context.Context, string) ([]installment.Installment, error) {
	return nil, nil
}
func (m *installmentMock) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]installment.Installment, error) {
	return nil, nil
}
func (m *installmentMock) DeleteByExpense(context.Context, string) error { return nil }

type recurringExpenseMock struct {
	items []recurringexpense.RecurringExpense
}

func (m *recurringExpenseMock) Save(context.Context, recurringexpense.RecurringExpense) error {
	return nil
}
func (m *recurringExpenseMock) FindByID(context.Context, string) (recurringexpense.RecurringExpense, error) {
	return recurringexpense.RecurringExpense{}, shared.ErrNotFound
}
func (m *recurringExpenseMock) List(_ context.Context, _ string, _, _ int) ([]recurringexpense.RecurringExpense, error) {
	return m.items, nil
}
func (m *recurringExpenseMock) ListDueForGeneration(context.Context, time.Time) ([]recurringexpense.RecurringExpense, error) {
	return nil, nil
}
func (m *recurringExpenseMock) Update(context.Context, recurringexpense.RecurringExpense) error {
	return nil
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
func (m *householdMock) ListByUserID(_ context.Context, _ string, _, _ int) ([]household.Household, error) {
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
func (m *memberMock) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *memberMock) FindByUserIDGlobal(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *memberMock) ListHouseholdIDsByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *memberMock) Update(context.Context, member.Member) error { return nil }
func (m *memberMock) Delete(_ context.Context, _ string) error    { return nil }
func (m *memberMock) CountActiveByHousehold(_ context.Context, _ string) (int, error) {
	return len(m.members), nil
}

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
func (m *expenseMock) SumPersonalByMemberAndPeriod(_ context.Context, _, _ string, _, _ time.Time) (int64, error) {
	return 0, nil
}
