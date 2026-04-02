package settlement

import (
	"errors"
	"sort"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/member"
)

var (
	ErrNoMembers = errors.New("no members in household")
)

// Transfer is a suggested movement to settle balances.
type Transfer struct {
	FromMemberID string
	ToMemberID   string
	AmountCents  int64
}

// MemberResult contains per-member settlement math.
type MemberResult struct {
	MemberID        string
	Name            string
	PaidCents       int64
	ExpectedShare   int64
	NetBalanceCents int64
	SalaryWeightBps int64
}

// Result is the settlement output for a period.
type Result struct {
	SettlementMode          household.SettlementMode
	EffectiveSettlementMode household.SettlementMode
	FallbackReason          string
	TotalSharedCents        int64
	IncludedExpenseCount    int
	ExcludedVoucherCount    int
	Members                 []MemberResult
	Transfers               []Transfer
}

// Calculate computes per-member balances and transfer suggestions.
func Calculate(mode household.SettlementMode, members []member.Member, expenses []expense.Expense, installments []installment.Installment) (Result, error) {
	if len(members) == 0 {
		return Result{}, ErrNoMembers
	}

	result := Result{SettlementMode: mode, EffectiveSettlementMode: mode}
	memberIndex := make(map[string]int, len(members))
	memberResults := make([]MemberResult, 0, len(members))

	for i, m := range members {
		memberID := string(m.ID())
		memberIndex[memberID] = i
		memberResults = append(memberResults, MemberResult{MemberID: memberID, Name: m.Name()})
	}

	for _, e := range expenses {
		if e.DeletedAt() != nil || !e.IsShared() {
			continue
		}

		// MSI root expenses are excluded because their value is settled via installments.
		if e.ExpenseType() == expense.ExpenseTypeMSI {
			continue
		}

		// Vouchers were previously excluded but now they are included in settlement.
		// Previous code checked for PaymentMethodVoucher and continued.

		result.TotalSharedCents += e.AmountCents()
		result.IncludedExpenseCount++

		idx, ok := memberIndex[e.PaidByMemberID()]
		if ok {
			memberResults[idx].PaidCents += e.AmountCents()
		}
	}

	for _, i := range installments {
		result.TotalSharedCents += i.InstallmentAmountCents()
		result.IncludedExpenseCount++

		idx, ok := memberIndex[i.PaidByMemberID()]
		if ok {
			memberResults[idx].PaidCents += i.InstallmentAmountCents()
		}
	}

	expectedShares, weights, effectiveMode, fallbackReason := computeExpectedShares(mode, members, result.TotalSharedCents)
	result.EffectiveSettlementMode = effectiveMode
	result.FallbackReason = fallbackReason

	for i := range memberResults {
		memberResults[i].ExpectedShare = expectedShares[i]
		memberResults[i].SalaryWeightBps = weights[i]
		memberResults[i].NetBalanceCents = memberResults[i].PaidCents - expectedShares[i]
	}

	result.Members = memberResults
	result.Transfers = suggestTransfers(memberResults)
	return result, nil
}

func computeExpectedShares(mode household.SettlementMode, members []member.Member, total int64) ([]int64, []int64, household.SettlementMode, string) {
	shares := make([]int64, len(members))
	weightsBps := make([]int64, len(members))

	if total <= 0 {
		return shares, weightsBps, mode, ""
	}

	if mode == household.SettlementModeProportional {
		salaryWeights := make([]int64, len(members))
		var salaryTotal int64
		for i, m := range members {
			salary := m.MonthlySalaryCents()
			salaryWeights[i] = salary
			salaryTotal += salary
		}

		if salaryTotal > 0 {
			shares = allocateByWeights(total, salaryWeights)
			for i := range salaryWeights {
				weightsBps[i] = (salaryWeights[i] * 10_000) / salaryTotal
			}
			return shares, weightsBps, household.SettlementModeProportional, ""
		}

		for i := range weightsBps {
			weightsBps[i] = int64(10_000 / len(members))
		}
		return allocateEqual(total, len(members)), weightsBps, household.SettlementModeEqual, "proportional mode fallback to equal because total salary is zero"
	}

	for i := range weightsBps {
		weightsBps[i] = int64(10_000 / len(members))
	}
	return allocateEqual(total, len(members)), weightsBps, household.SettlementModeEqual, ""
}

func allocateEqual(total int64, n int) []int64 {
	shares := make([]int64, n)
	if n == 0 {
		return shares
	}
	base := total / int64(n)
	rem := total % int64(n)
	for i := 0; i < n; i++ {
		shares[i] = base
		if int64(i) < rem {
			shares[i]++
		}
	}
	return shares
}

func allocateByWeights(total int64, weights []int64) []int64 {
	type remainderItem struct {
		idx       int
		remainder int64
	}

	shares := make([]int64, len(weights))
	var weightTotal int64
	for _, w := range weights {
		weightTotal += w
	}
	if weightTotal <= 0 {
		return shares
	}

	allocated := int64(0)
	remainders := make([]remainderItem, 0, len(weights))
	for i, w := range weights {
		numerator := total * w
		shares[i] = numerator / weightTotal
		allocated += shares[i]
		remainders = append(remainders, remainderItem{idx: i, remainder: numerator % weightTotal})
	}

	sort.SliceStable(remainders, func(i, j int) bool {
		return remainders[i].remainder > remainders[j].remainder
	})

	left := total - allocated
	for i := int64(0); i < left; i++ {
		shares[remainders[i].idx]++
	}

	return shares
}

func suggestTransfers(results []MemberResult) []Transfer {
	type bucket struct {
		memberID string
		amount   int64
	}

	creditors := make([]bucket, 0)
	debtors := make([]bucket, 0)
	for _, r := range results {
		if r.NetBalanceCents > 0 {
			creditors = append(creditors, bucket{memberID: r.MemberID, amount: r.NetBalanceCents})
		}
		if r.NetBalanceCents < 0 {
			debtors = append(debtors, bucket{memberID: r.MemberID, amount: -r.NetBalanceCents})
		}
	}

	sort.SliceStable(creditors, func(i, j int) bool { return creditors[i].amount > creditors[j].amount })
	sort.SliceStable(debtors, func(i, j int) bool { return debtors[i].amount > debtors[j].amount })

	transfers := make([]Transfer, 0)
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		pay := debtors[i].amount
		if creditors[j].amount < pay {
			pay = creditors[j].amount
		}
		if pay > 0 {
			transfers = append(transfers, Transfer{
				FromMemberID: debtors[i].memberID,
				ToMemberID:   creditors[j].memberID,
				AmountCents:  pay,
			})
		}

		debtors[i].amount -= pay
		creditors[j].amount -= pay
		if debtors[i].amount == 0 {
			i++
		}
		if creditors[j].amount == 0 {
			j++
		}
	}

	return transfers
}
