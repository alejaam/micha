package settlement_test

import (
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
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

	res, err := settlement.Calculate(household.SettlementModeEqual, members, expenses, nil)
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

func TestCalculate_IncludeVoucherAndMSIInstallments(t *testing.T) {
	t.Parallel()
	members := []member.Member{
		mustMember(t, "m-1", "hh-1", "Ana", 300000),
		mustMember(t, "m-2", "hh-1", "Luis", 100000),
	}
	expenses := []expense.Expense{
		mustExpense(t, expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-1", AmountCents: 40000, Description: "rent", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodTransfer, CreatedAt: time.Now()}),
		// Voucher should be included now
		mustExpense(t, expense.ExpenseAttributes{ID: "e-2", HouseholdID: "hh-1", PaidByMemberID: "m-2", AmountCents: 10000, Description: "voucher purchase", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodVoucher, CreatedAt: time.Now()}),
		// MSI root should be EXCLUDED
		mustExpense(t, expense.ExpenseAttributes{ID: "e-3", HouseholdID: "hh-1", PaidByMemberID: "m-1", AmountCents: 90000, Description: "iphone", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodCard, ExpenseType: expense.ExpenseTypeMSI, TotalInstallments: 3, CreatedAt: time.Now()}),
	}
	// Add installment for the MSI root
	installments := []installment.Installment{
		mustInstallment(t, "i-1", "e-3", "m-1", 30000, 90000, 3, 1),
	}

	res, err := settlement.Calculate(household.SettlementModeProportional, members, expenses, installments)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Total = 40000 (rent) + 10000 (voucher) + 30000 (installment) = 80000
	if res.TotalSharedCents != 80000 {
		t.Fatalf("total shared = %d, want 80000", res.TotalSharedCents)
	}

	// Ana (m-1) paid: 40000 (rent) + 30000 (inst) = 70000
	// Luis (m-2) paid: 10000 (voucher)
	// Salaries: Ana 300k, Luis 100k -> 3:1 ratio
	// Expected shares: Ana 60000, Luis 20000
	// Balances: Ana +10000, Luis -10000
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

	res, err := settlement.Calculate(household.SettlementModeProportional, members, expenses, nil)
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

func TestApplyAdditionalShared_RebalancesWithoutPayer(t *testing.T) {
	t.Parallel()
	members := []member.Member{
		mustMember(t, "m-1", "hh-1", "Ana", 300000),
		mustMember(t, "m-2", "hh-1", "Luis", 100000),
	}
	expenses := []expense.Expense{
		mustExpense(t, expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-1", AmountCents: 10000, Description: "food", IsShared: true, Currency: "MXN", PaymentMethod: expense.PaymentMethodCash, CreatedAt: time.Now()}),
	}

	res, err := settlement.Calculate(household.SettlementModeProportional, members, expenses, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Add 20,000 shared fixed agnostic amount (no payer).
	res = settlement.ApplyAdditionalShared(res, household.SettlementModeProportional, members, 20000, 1)

	// Total = 30,000. Proportional 3:1 -> expected: 22,500 / 7,500.
	if res.TotalSharedCents != 30000 {
		t.Fatalf("total shared = %d, want 30000", res.TotalSharedCents)
	}
	// No payer exists for the agnostic amount, so balances can be negative for all
	// members and there should be no internal transfer suggestions.
	if len(res.Transfers) != 0 {
		t.Fatalf("transfers = %d, want 0", len(res.Transfers))
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

func mustInstallment(t *testing.T, id, expenseID, memberID string, amount, total int64, count, current int) installment.Installment {
	t.Helper()
	i, err := installment.New(installment.ID(id), expenseID, memberID, time.Now(), amount, total, count, current, time.Now())
	if err != nil {
		t.Fatalf("installment.New: %v", err)
	}
	return i
}

func TestCalculate_ChargedToMember(t *testing.T) {
	t.Parallel()
	exp := func(id, paidBy, chargedTo string, amount int64) expense.Expense {
		return mustExpense(t, expense.ExpenseAttributes{ID: expense.ID(id), HouseholdID: "hh-1", PaidByMemberID: paidBy, ChargedToMemberID: chargedTo, AmountCents: amount, Description: id, IsShared: true, Currency: "MXN", CreatedAt: time.Now()})
	}
	cases := []struct {
		name         string
		mode         household.SettlementMode
		salaryA      int64
		expenses     []expense.Expense
		wantShared   int64
		wantFallback int
		wantNetB     int64
		want         []settlement.Transfer
	}{
		{"acceptance 50/50", household.SettlementModeEqual, 100000,
			[]expense.Expense{exp("e-1", "m-a", "", 600892), exp("e-2", "m-a", "m-b", 142500)},
			600892, 0, -442946, []settlement.Transfer{{FromMemberID: "m-b", ToMemberID: "m-a", AmountCents: 442946}}},
		{"proportional not split", household.SettlementModeProportional, 300000,
			[]expense.Expense{exp("e-1", "m-b", "m-a", 5000)},
			0, 0, 5000, []settlement.Transfer{{FromMemberID: "m-a", ToMemberID: "m-b", AmountCents: 5000}}},
		{"charged to payer", household.SettlementModeEqual, 100000,
			[]expense.Expense{exp("e-1", "m-a", "m-a", 142500)}, 0, 0, 0, nil},
		{"removed assigned member falls back to shared and is flagged", household.SettlementModeEqual, 100000,
			[]expense.Expense{exp("e-1", "m-a", "m-gone", 10000)},
			10000, 1, -5000, []settlement.Transfer{{FromMemberID: "m-b", ToMemberID: "m-a", AmountCents: 5000}}},
		// Payer no longer active: the explicit mark is honoured (B still owes the
		// full amount); there is no active member to credit, so no transfer.
		{"removed payer keeps full charge on assigned member", household.SettlementModeEqual, 100000,
			[]expense.Expense{exp("e-1", "m-gone", "m-b", 10000)},
			0, 0, -10000, nil},
		// A non-shared expense with a removed assigned member is excluded, so it is
		// not reported as a fallback to shared (review of PR #85).
		{"non-shared with removed assigned member is not a fallback", household.SettlementModeEqual, 100000,
			[]expense.Expense{mustExpense(t, expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-a", ChargedToMemberID: "m-gone", AmountCents: 10000, Description: "e-1", IsShared: false, Currency: "MXN", CreatedAt: time.Now()})},
			0, 0, 0, nil},
	}
	for _, tc := range cases {
		members := []member.Member{mustMember(t, "m-a", "hh-1", "A", tc.salaryA), mustMember(t, "m-b", "hh-1", "B", 100000)}
		res, err := settlement.Calculate(tc.mode, members, tc.expenses, nil)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if res.TotalSharedCents != tc.wantShared {
			t.Fatalf("%s: total shared = %d, want %d", tc.name, res.TotalSharedCents, tc.wantShared)
		}
		if res.ChargedFallbackCount != tc.wantFallback {
			t.Fatalf("%s: charged fallback count = %d, want %d", tc.name, res.ChargedFallbackCount, tc.wantFallback)
		}
		if res.Members[1].NetBalanceCents != tc.wantNetB {
			t.Fatalf("%s: net balance B = %d, want %d", tc.name, res.Members[1].NetBalanceCents, tc.wantNetB)
		}
		if len(res.Transfers) != len(tc.want) || (len(tc.want) == 1 && res.Transfers[0] != tc.want[0]) {
			t.Fatalf("%s: transfers = %+v, want %+v", tc.name, res.Transfers, tc.want)
		}
	}
}

func TestApplyAdditionalShared_PreservesCharges(t *testing.T) {
	t.Parallel()
	members := []member.Member{mustMember(t, "m-a", "hh-1", "A", 100000), mustMember(t, "m-b", "hh-1", "B", 100000)}
	expenses := []expense.Expense{
		mustExpense(t, expense.ExpenseAttributes{ID: "e-1", HouseholdID: "hh-1", PaidByMemberID: "m-a", ChargedToMemberID: "m-b", AmountCents: 1000, Description: "casco", IsShared: true, Currency: "MXN", CreatedAt: time.Now()}),
	}
	res, err := settlement.Calculate(household.SettlementModeEqual, members, expenses, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res = settlement.ApplyAdditionalShared(res, household.SettlementModeEqual, members, 2000, 1)
	// Shared 2000 unpaid → each owes 1000; B also owes 1000 to A.
	if res.Members[0].NetBalanceCents != 0 || res.Members[1].NetBalanceCents != -2000 {
		t.Fatalf("unexpected balances: %+v", res.Members)
	}
}
