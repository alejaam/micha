package periodapp

import (
	"context"
	"fmt"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/period"
	periodapproval "micha/backend/internal/domain/period_approval"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type ClosePeriodUseCase struct {
	periodRepo      outbound.PeriodRepository
	approvalRepo    outbound.PeriodApprovalRepository
	householdRepo   outbound.HouseholdRepository
	memberRepo      outbound.MemberRepository
	expenseRepo     outbound.ExpenseRepository
	installmentRepo outbound.InstallmentRepository
	txManager       outbound.TransactionManager
	idGenerator     appshared.IDGenerator
	now             func() time.Time
}

func NewClosePeriodUseCase(
	periodRepo outbound.PeriodRepository,
	approvalRepo outbound.PeriodApprovalRepository,
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	expenseRepo outbound.ExpenseRepository,
	installmentRepo outbound.InstallmentRepository,
	txManager outbound.TransactionManager,
	idGenerator appshared.IDGenerator,
) ClosePeriodUseCase {
	return ClosePeriodUseCase{
		periodRepo:      periodRepo,
		approvalRepo:    approvalRepo,
		householdRepo:   householdRepo,
		memberRepo:      memberRepo,
		expenseRepo:     expenseRepo,
		installmentRepo: installmentRepo,
		txManager:       txManager,
		idGenerator:     idGenerator,
		now:             appshared.Now,
	}
}

func (u ClosePeriodUseCase) Execute(ctx context.Context, input inbound.ClosePeriodInput) (inbound.ClosePeriodOutput, error) {
	// 1. Retrieve the household and current member/actor.
	h, err := u.householdRepo.FindByID(ctx, input.HouseholdID)
	if err != nil {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: %w", err)
	}

	if _, findErr := u.memberRepo.FindByUserID(ctx, input.HouseholdID, input.CurrentUserID); findErr != nil {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: %w", findErr)
	}

	// 2. Retrieve the period and validate status.
	p, err := u.periodRepo.GetByID(ctx, period.ID(input.PeriodID))
	if err != nil {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: %w", err)
	}

	if p.Status() != period.StatusReview {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: only periods in review can be closed")
	}

	// 3. Only the household owner can close a period.
	if h.OwnerID() != input.CurrentUserID {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: %w", shared.ErrForbidden)
	}

	// 3.5. Minimum duration guard: a period must last at least 7 days before it can be closed.
	now := u.now()
	if now.Sub(p.StartDate()) < period.MinimumPeriodDuration {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: %w", period.ErrPeriodTooShort)
	}

	// 3.6. Future-period guard: reject if the next period would start in the future.
	nextStart := p.EndDate().Add(24 * time.Hour)
	nextStartNorm := time.Date(nextStart.Year(), nextStart.Month(), nextStart.Day(), 0, 0, 0, 0, time.UTC)
	todayNorm := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if nextStartNorm.After(todayNorm) {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: %w", shared.ErrFuturePeriod)
	}

	// 4. Consensus Check (non-force requires unanimous approval).
	if !input.Force {
		if err := u.validateConsensus(ctx, input.HouseholdID, input.PeriodID); err != nil {
			return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: consensus required: %w", err)
		}
	}

	// 4-7. Execute transactional operations (close period, create next, rollovers).
	var nextPeriodID string
	if err := u.txManager.Run(ctx, func(txCtx context.Context) error {
		pAttrs := p.Attributes()
		pAttrs.Status = period.StatusClosed
		pAttrs.UpdatedAt = now
		closedPeriod, _ := period.NewFromAttributes(pAttrs)

		if err := u.periodRepo.Update(txCtx, closedPeriod); err != nil {
			return fmt.Errorf("failed to close: %w", err)
		}

		// 5. Create Rollover (Next Period).
		var nextEnd time.Time
		if h.Attributes().PeriodFrequency == "biweekly" {
			// Biweekly: 1-15 and 16-last-day-of-month
			if nextStart.Day() == 1 {
				nextEnd = time.Date(nextStart.Year(), nextStart.Month(), 15, 23, 59, 59, 999999999, nextStart.Location())
			} else {
				// nextStart is 16th → end is last day of month
				nextEnd = time.Date(nextStart.Year(), nextStart.Month()+1, 0, 23, 59, 59, 999999999, nextStart.Location())
			}
		} else {
			// Monthly: closingDay of next month
			closingDay := h.Attributes().ClosingDay
			nextEnd = time.Date(nextStart.Year(), nextStart.Month()+1, closingDay, 23, 59, 59, 999999999, nextStart.Location())
		}

		nextPeriod, err := period.New(
			period.ID(u.idGenerator.NewID()),
			input.HouseholdID,
			nextStart,
			nextEnd,
			period.StatusOpen,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to create next: %w", err)
		}

		if err := u.periodRepo.Create(txCtx, nextPeriod); err != nil {
			return fmt.Errorf("failed to persist next: %w", err)
		}

		// 6. Rollover Fixed Expenses.
		if err := u.rolloverFixedExpenses(txCtx, input.PeriodID, string(nextPeriod.ID()), now); err != nil {
			return fmt.Errorf("fixed rollover failed: %w", err)
		}

		// 7. Rollover Installments (MSI).
		if err := u.rolloverInstallments(txCtx, nextPeriod, now); err != nil {
			return fmt.Errorf("MSI rollover failed: %w", err)
		}

		nextPeriodID = string(nextPeriod.ID())
		return nil
	}); err != nil {
		return inbound.ClosePeriodOutput{}, fmt.Errorf("close period: %w", err)
	}

	return inbound.ClosePeriodOutput{NextPeriodID: nextPeriodID}, nil
}

func (u ClosePeriodUseCase) validateConsensus(ctx context.Context, householdID, periodID string) error {
	members, err := u.memberRepo.ListAllByHousehold(ctx, householdID)
	if err != nil {
		return err
	}

	approvals, err := u.approvalRepo.ListByPeriod(ctx, periodID)
	if err != nil {
		return err
	}

	approvalMap := make(map[string]periodapproval.ApprovalStatus)
	for _, a := range approvals {
		approvalMap[a.MemberID()] = a.Status()
	}

	for _, m := range members {
		status, exists := approvalMap[string(m.ID())]
		if !exists {
			return fmt.Errorf("member %s has not voted", m.ID())
		}
		if status == periodapproval.ApprovalStatusObjected {
			return fmt.Errorf("member %s has objected", m.ID())
		}
	}

	return nil
}

func (u ClosePeriodUseCase) rolloverFixedExpenses(ctx context.Context, currentPeriodID, nextPeriodID string, now time.Time) error {
	expenses, err := u.expenseRepo.ListByPeriod(ctx, currentPeriodID)
	if err != nil {
		return err
	}

	for _, e := range expenses {
		if e.ExpenseType() == expense.ExpenseTypeFixed {
			attrs := e.Attributes()
			attrs.ID = expense.ID(u.idGenerator.NewID())
			attrs.PeriodID = nextPeriodID
			attrs.CreatedAt = now
			attrs.UpdatedAt = now

			cloned, err := expense.NewFromAttributes(attrs)
			if err != nil {
				return fmt.Errorf("failed to clone fixed expense: %w", err)
			}
			if err := u.expenseRepo.Save(ctx, cloned); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u ClosePeriodUseCase) rolloverInstallments(ctx context.Context, nextPeriod period.Period, now time.Time) error {
	// Find installments whose StartDate falls within the next period.
	// BUT, our Expense entity now has period_id. For each installment due in the next period,
	// we should probably create an Expense record of type 'msi' linked to that period.

	installments, err := u.installmentRepo.ListByHouseholdAndPeriod(ctx, nextPeriod.HouseholdID(), nextPeriod.StartDate(), nextPeriod.EndDate())
	if err != nil {
		return err
	}

	for _, inst := range installments {
		// Create a virtual expense for this installment in the new period.
		// This makes the installment visible in the expense list for the month.
		// We need the parent expense's category_id to satisfy the NOT NULL FK constraint.
		parentExpense, err := u.expenseRepo.FindByID(ctx, inst.ExpenseID())
		if err != nil {
			return fmt.Errorf("failed to find parent expense for installment %s: %w", inst.ExpenseID(), err)
		}

		e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
			ID:                expense.ID(u.idGenerator.NewID()),
			HouseholdID:       nextPeriod.HouseholdID(),
			PaidByMemberID:    inst.PaidByMemberID(),
			PeriodID:          string(nextPeriod.ID()),
			CategoryID:        parentExpense.CategoryID(),
			AmountCents:       inst.InstallmentAmountCents(),
			Description:       fmt.Sprintf("MSI installment %d/%d", inst.CurrentInstallment(), inst.TotalInstallments()),
			IsShared:          true, // MSI root expense defines this, but for simplicity...
			Currency:          "MXN",
			PaymentMethod:     expense.PaymentMethodCard,
			ExpenseType:       expense.ExpenseTypeMSI,
			TotalInstallments: inst.TotalInstallments(),
			CreatedAt:         now,
			UpdatedAt:         now,
		})
		if err != nil {
			return fmt.Errorf("failed to create MSI expense for installment: %w", err)
		}
		if err := u.expenseRepo.Save(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

var _ inbound.ClosePeriodUseCase = ClosePeriodUseCase{}
