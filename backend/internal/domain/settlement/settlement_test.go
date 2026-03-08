package settlement_test

import (
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/settlement"
)

func TestCalculate_EqualMode(t *testing.T) {
	t.Parallel()
	members := []member.Member{
		mustMember(t, "m-1", "hh-1", "Ana", 300000),
		mustMember(t, "m-2", "hh-1", "Luis", 100000),
	}
	expenses := []expense.Expense{
		mustExpense(t, expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-1", AmountCents: 10000, Description: "food", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodCard, CreatedAt: time.Now()}),
		mustExpense(t, expense.ExpenseAttributes{ID: "e-2", HouseholdID: "hh-1", PaidByMemberID: "m-2", AmountCents: 5000, Description: "uber", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodCash, CreatedAt: time.Now()}),
	}

	res, err := settlement.Calculate(household.SettlementModeEqual, members, expenses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalSharedCents != 15000 {
		t.Fatalf("total shared = %d, want 15000", res.TotalSharedCents)
	}
	if len(res.Transfers) != 1 {
		t.Fatalf("transfers = %d, want 1", len(res.Transfers))
	}
	if res.Transfers[0].FromMemberID != "m-2" || res.Transfers[0].ToMemberID != "m-1" || res.Transfers[0].AmountCents != 2500 {
		t.Fatalf("unexpected transfer: %+v", res.Transfers[0])
	}
}

func TestCalculate_ProportionalModeWithVoucherExclusion(t *testing.T) {
	t.Parallel()
	members := []member.Member{
		mustMember(t, "m-1", "hh-1", "Ana", 300000),
		mustMember(t, "m-2", "hh-1", "Luis", 100000),
	}
	expenses := []expense.Expense{
		mustExpense(t, expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-1", AmountCents: 40000, Description: "rent", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodTransfer, CreatedAt: time.Now()}),
		mustExpense(t, expense.ExpenseAttributes{ID: "e-2", HouseholdID: "hh-1", PaidByMemberID: "m-2", AmountCents: 10000, Description: "voucher purchase", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodVoucher, CreatedAt: time.Now()}),
	}

	res, err := settlement.Calculate(household.SettlementModeProportional, members, expenses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalSharedCents != 40000 {
		t.Fatalf("total shared = %d, want 40000", res.TotalSharedCents)
	}
	if res.ExcludedVoucherCount != 1 {
		t.Fatalf("excluded vouchers = %d, want 1", res.ExcludedVoucherCount)
	}
	if len(res.Transfers) != 1 {
		t.Fatalf("transfers = %d, want 1", len(res.Transfers))
	}
	if res.Transfers[0].FromMemberID != "m-2" || res.Transfers[0].ToMemberID != "m-1" || res.Transfers[0].AmountCents != 10000 {
		t.Fatalf("unexpected transfer: %+v", res.Transfers[0])
	}
}

func TestCalculate_ProportionalFallsBackToEqualWhenSalaryIsZero(t *testing.T) {
	t.Parallel()
	members := []member.Member{
		mustMember(t, "m-1", "hh-1", "Ana", 0),
		mustMember(t, "m-2", "hh-1", "Luis", 0),
	}
	expenses := []expense.Expense{
		mustExpense(t, expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-1", AmountCents: 10000, Description: "food", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodCash, CreatedAt: time.Now()}),
	}

	res, err := settlement.Calculate(household.SettlementModeProportional, members, expenses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.EffectiveSettlementMode != household.SettlementModeEqual {
		t.Fatalf("effective mode = %s, want equal", res.EffectiveSettlementMode)
	}
	if res.FallbackReason == "" {
		t.Fatal("expected fallback reason")
	}
}

func mustMember(t *testing.T, id, householdID, name string, salary int64) member.Member {
	t.Helper()
	m, err := member.New(member.ID(id), householdID, name, name+"@mail.com", salary, time.Now())
	if err != nil {
		t.Fatalf("member.New: %v", err)
	}
	return m
}

func mustExpense(t *testing.T, attrs expense.ExpenseAttributes) expense.Expense {
	t.Helper()
	e, err := expense.NewFromAttributes(attrs)
	if err != nil {
		t.Fatalf("expense.NewFromAttributes: %v", err)
	}
	return e
}
