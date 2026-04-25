package memberapp

import (
	"context"
	"fmt"
	"time"

	"micha/backend/internal/domain/settlement"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// CalculateRemainingSalaryUseCase computes period remaining salary for one member.
type CalculateRemainingSalaryUseCase struct {
	householdRepo   outbound.HouseholdRepository
	memberRepo      outbound.MemberRepository
	expenseRepo     outbound.ExpenseRepository
	installmentRepo outbound.InstallmentRepository
}

// NewCalculateRemainingSalaryUseCase builds CalculateRemainingSalaryUseCase.
func NewCalculateRemainingSalaryUseCase(
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	expenseRepo outbound.ExpenseRepository,
	installmentRepo outbound.InstallmentRepository,
) CalculateRemainingSalaryUseCase {
	return CalculateRemainingSalaryUseCase{
		householdRepo:   householdRepo,
		memberRepo:      memberRepo,
		expenseRepo:     expenseRepo,
		installmentRepo: installmentRepo,
	}
}

// Execute returns monthly salary, personal outflow, shared allocated debt and remaining salary.
func (u CalculateRemainingSalaryUseCase) Execute(ctx context.Context, input inbound.CalculateRemainingSalaryInput) (inbound.CalculateRemainingSalaryOutput, error) {
	if input.HouseholdID == "" || input.MemberID == "" {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: %w", shared.ErrInvalidID)
	}
	if input.Year < 2000 || input.Year > 2200 {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: invalid year")
	}
	if input.Month < 1 || input.Month > 12 {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: invalid month")
	}

	m, err := u.memberRepo.FindByID(ctx, input.MemberID)
	if err != nil {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: find member: %w", err)
	}
	if m.HouseholdID() != input.HouseholdID {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: %w", shared.ErrForbidden)
	}

	from := time.Date(input.Year, time.Month(input.Month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	personalOutflow, err := u.expenseRepo.SumPersonalByMemberAndPeriod(ctx, input.HouseholdID, input.MemberID, from, to)
	if err != nil {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: personal outflow: %w", err)
	}

	householdEntity, err := u.householdRepo.FindByID(ctx, input.HouseholdID)
	if err != nil {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: find household: %w", err)
	}

	members, err := u.memberRepo.ListAllByHousehold(ctx, input.HouseholdID)
	if err != nil {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: list members: %w", err)
	}

	expenses, err := u.expenseRepo.ListByHouseholdAndPeriod(ctx, input.HouseholdID, from, to)
	if err != nil {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: list expenses: %w", err)
	}

	installments, err := u.installmentRepo.ListByHouseholdAndPeriod(ctx, input.HouseholdID, from, to)
	if err != nil {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: list installments: %w", err)
	}

	calc, err := settlement.Calculate(householdEntity.SettlementMode(), members, expenses, installments)
	if err != nil {
		return inbound.CalculateRemainingSalaryOutput{}, fmt.Errorf("calculate remaining salary: settlement calculate: %w", err)
	}

	allocatedDebt := int64(0)
	for _, memberResult := range calc.Members {
		if memberResult.MemberID == input.MemberID {
			allocatedDebt = memberResult.ExpectedShare
			break
		}
	}

	remainingSalary := m.MonthlySalaryCents() - personalOutflow - allocatedDebt

	return inbound.CalculateRemainingSalaryOutput{
		HouseholdID:           input.HouseholdID,
		MemberID:              input.MemberID,
		Year:                  input.Year,
		Month:                 input.Month,
		MonthlySalaryCents:    m.MonthlySalaryCents(),
		PersonalExpensesCents: personalOutflow,
		AllocatedDebtCents:    allocatedDebt,
		RemainingSalaryCents:  remainingSalary,
	}, nil
}

var _ inbound.CalculateRemainingSalaryUseCase = CalculateRemainingSalaryUseCase{}
