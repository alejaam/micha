package settlementapp

import (
	"context"
	"fmt"
	"time"

	"micha/backend/internal/domain/settlement"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

var _ inbound.CalculateSettlementUseCase = CalculateSettlementUseCase{}

// CalculateSettlementUseCase computes a household monthly settlement report.
type CalculateSettlementUseCase struct {
	householdRepo   outbound.HouseholdRepository
	memberRepo      outbound.MemberRepository
	expenseRepo     outbound.ExpenseRepository
	installmentRepo outbound.InstallmentRepository
}

// NewCalculateSettlementUseCase constructs CalculateSettlementUseCase.
func NewCalculateSettlementUseCase(
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	expenseRepo outbound.ExpenseRepository,
	installmentRepo outbound.InstallmentRepository,
) CalculateSettlementUseCase {
	return CalculateSettlementUseCase{
		householdRepo:   householdRepo,
		memberRepo:      memberRepo,
		expenseRepo:     expenseRepo,
		installmentRepo: installmentRepo,
	}
}

// Execute calculates monthly settlement for one household.
func (u CalculateSettlementUseCase) Execute(ctx context.Context, input inbound.CalculateSettlementInput) (inbound.CalculateSettlementOutput, error) {
	if input.HouseholdID == "" {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: %w", inbound.ErrSettlementMissingHouseholdID)
	}
	if input.Year < 2000 || input.Year > 2200 {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: %w", inbound.ErrSettlementYearOutOfRange)
	}
	if input.Month < 1 || input.Month > 12 {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: %w", inbound.ErrSettlementInvalidMonth)
	}

	householdEntity, err := u.householdRepo.FindByID(ctx, input.HouseholdID)
	if err != nil {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: find household: %w", err)
	}

	members, err := u.memberRepo.ListAllByHousehold(ctx, input.HouseholdID)
	if err != nil {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: list members: %w", err)
	}

	from := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	expenses, err := u.expenseRepo.ListByHouseholdAndPeriod(ctx, input.HouseholdID, from, to)
	if err != nil {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: list expenses: %w", err)
	}

	installments, err := u.installmentRepo.ListByHouseholdAndPeriod(ctx, input.HouseholdID, from, to)
	if err != nil {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: list installments: %w", err)
	}

	calc, err := settlement.Calculate(householdEntity.SettlementMode(), members, expenses, installments)
	if err != nil {
		return inbound.CalculateSettlementOutput{}, fmt.Errorf("calculate settlement: %w", err)
	}

	memberOut := make([]inbound.MemberSettlement, 0, len(calc.Members))
	for _, m := range calc.Members {
		memberOut = append(memberOut, inbound.MemberSettlement{
			MemberID:        m.MemberID,
			Name:            m.Name,
			PaidCents:       m.PaidCents,
			ExpectedShare:   m.ExpectedShare,
			NetBalanceCents: m.NetBalanceCents,
			SalaryWeightBps: m.SalaryWeightBps,
		})
	}

	transferOut := make([]inbound.SettlementTransfer, 0, len(calc.Transfers))
	for _, t := range calc.Transfers {
		transferOut = append(transferOut, inbound.SettlementTransfer{
			FromMemberID: t.FromMemberID,
			ToMemberID:   t.ToMemberID,
			AmountCents:  t.AmountCents,
		})
	}

	return inbound.CalculateSettlementOutput{
		HouseholdID:             input.HouseholdID,
		Year:                    input.Year,
		Month:                   input.Month,
		SettlementMode:          calc.SettlementMode,
		EffectiveSettlementMode: calc.EffectiveSettlementMode,
		FallbackReason:          calc.FallbackReason,
		TotalSharedCents:        calc.TotalSharedCents,
		IncludedExpenseCount:    calc.IncludedExpenseCount,
		ExcludedVoucherCount:    calc.ExcludedVoucherCount,
		Members:                 memberOut,
		Transfers:               transferOut,
	}, nil
}
