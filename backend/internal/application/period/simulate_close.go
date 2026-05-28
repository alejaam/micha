package periodapp

import (
	"context"
	"fmt"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// SimulateClosePeriodUseCase produces a read-only projection of what would happen
// if the period were closed: next period dates, settlement preview, and rollover counts.
// It does NOT persist any changes.
type SimulateClosePeriodUseCase struct {
	periodRepo      outbound.PeriodRepository
	householdRepo   outbound.HouseholdRepository
	memberRepo      outbound.MemberRepository
	expenseRepo     outbound.ExpenseRepository
	installmentRepo outbound.InstallmentRepository
	now             func() time.Time
}

// NewSimulateClosePeriodUseCase constructs the use case.
func NewSimulateClosePeriodUseCase(
	periodRepo outbound.PeriodRepository,
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	expenseRepo outbound.ExpenseRepository,
	installmentRepo outbound.InstallmentRepository,
) SimulateClosePeriodUseCase {
	return SimulateClosePeriodUseCase{
		periodRepo:      periodRepo,
		householdRepo:   householdRepo,
		memberRepo:      memberRepo,
		expenseRepo:     expenseRepo,
		installmentRepo: installmentRepo,
		now:             appshared.Now,
	}
}

// Execute produces a projection of what would happen if the period were closed.
func (u SimulateClosePeriodUseCase) Execute(ctx context.Context, input inbound.SimulateClosePeriodInput) (inbound.SimulateClosePeriodOutput, error) {
	// 1. Validate ownership and membership.
	h, err := u.householdRepo.FindByID(ctx, input.HouseholdID)
	if err != nil {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", err)
	}

	if h.OwnerID() != input.CurrentUserID {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", shared.ErrForbidden)
	}

	if _, findErr := u.memberRepo.FindByUserID(ctx, input.HouseholdID, input.CurrentUserID); findErr != nil {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", findErr)
	}

	// 2. Retrieve the period and validate.
	p, err := u.periodRepo.GetByID(ctx, period.ID(input.PeriodID))
	if err != nil {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", err)
	}

	if p.HouseholdID() != input.HouseholdID {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: period not found in household")
	}

	if p.Status() != period.StatusOpen {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: only open periods can be closed")
	}

	// 3. Minimum duration guard.
	now := u.now()
	if now.Sub(p.StartDate()) < period.MinimumPeriodDuration {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", period.ErrPeriodTooShort)
	}

	// 4. Future-period guard.
	nextStart := p.EndDate().Add(24 * time.Hour)
	nextStartNorm := time.Date(nextStart.Year(), nextStart.Month(), nextStart.Day(), 0, 0, 0, 0, time.UTC)
	todayNorm := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if nextStartNorm.After(todayNorm) {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", shared.ErrFuturePeriod)
	}

	// 5. Compute next period dates.
	var nextEnd time.Time
	if h.Attributes().PeriodFrequency == "biweekly" {
		if nextStart.Day() == 1 {
			nextEnd = time.Date(nextStart.Year(), nextStart.Month(), 15, 23, 59, 59, 999999999, nextStart.Location())
		} else {
			nextEnd = time.Date(nextStart.Year(), nextStart.Month()+1, 0, 23, 59, 59, 999999999, nextStart.Location())
		}
	} else {
		closingDay := h.Attributes().ClosingDay
		nextEnd = time.Date(nextStart.Year(), nextStart.Month()+1, closingDay, 23, 59, 59, 999999999, nextStart.Location())
	}

	// 6. Count fixed expenses and installments for rollover projection.
	fixedExpenses, err := u.expenseRepo.ListByPeriod(ctx, input.PeriodID)
	if err != nil {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", err)
	}

	fixedCount := 0
	for _, e := range fixedExpenses {
		if e.ExpenseType() == expense.ExpenseTypeFixed {
			fixedCount++
		}
	}

	installments, err := u.installmentRepo.ListByHouseholdAndPeriod(ctx, input.HouseholdID, nextStart, nextEnd)
	if err != nil {
		return inbound.SimulateClosePeriodOutput{}, fmt.Errorf("simulate close: %w", err)
	}

	// 7. Build settlement preview (simplified: just the list of members and their expected shares).
	// For now, return empty settlement preview — the real settlement calculation is complex and
	// belongs in the settlement service. This projection shows what's available.
	settlementPreview := []inbound.SettlementEntry{}

	return inbound.SimulateClosePeriodOutput{
		NextPeriodStart:   nextStart,
		NextPeriodEnd:     nextEnd,
		SettlementPreview: settlementPreview,
		FixedExpenseCount: fixedCount,
		InstallmentCount:  len(installments),
	}, nil
}

var _ inbound.SimulateClosePeriodUseCase = SimulateClosePeriodUseCase{}
