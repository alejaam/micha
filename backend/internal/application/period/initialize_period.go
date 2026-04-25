package periodapp

import (
	"context"
	"fmt"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type InitializePeriodUseCase struct {
	periodRepo    outbound.PeriodRepository
	householdRepo outbound.HouseholdRepository
	memberRepo    outbound.MemberRepository
	expenseRepo   outbound.ExpenseRepository
	idGenerator   appshared.IDGenerator
	now           func() time.Time
}

func NewInitializePeriodUseCase(
	periodRepo outbound.PeriodRepository,
	householdRepo outbound.HouseholdRepository,
	memberRepo outbound.MemberRepository,
	expenseRepo outbound.ExpenseRepository,
	idGenerator appshared.IDGenerator,
) InitializePeriodUseCase {
	return InitializePeriodUseCase{
		periodRepo:    periodRepo,
		householdRepo: householdRepo,
		memberRepo:    memberRepo,
		expenseRepo:   expenseRepo,
		idGenerator:   idGenerator,
		now:           time.Now,
	}
}

func (u InitializePeriodUseCase) Execute(ctx context.Context, input inbound.InitializePeriodInput) (inbound.InitializePeriodOutput, error) {
	// 1. Validate household and permissions.
	h, err := u.householdRepo.FindByID(ctx, input.HouseholdID)
	if err != nil {
		return inbound.InitializePeriodOutput{}, fmt.Errorf("initialize period: %w", err)
	}

	// Permissive check: if household has no owner (legacy), anyone can initialize.
	// If it has an owner, only the owner can do it.
	if h.OwnerID() != "" && h.OwnerID() != input.CurrentUserID {
		return inbound.InitializePeriodOutput{}, fmt.Errorf("initialize period: only owner can initialize: %w", shared.ErrForbidden)
	}

	// 2. Check if any period already exists.
	_, err = u.periodRepo.GetLatestByHousehold(ctx, input.HouseholdID)
	if err == nil {
		return inbound.InitializePeriodOutput{}, fmt.Errorf("initialize period: household already has periods")
	}

	// 3. Create initial period (from start of current month to end of month).
	now := u.now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, -1)

	p, err := period.New(
		period.ID(u.idGenerator.NewID()),
		input.HouseholdID,
		start,
		end,
		period.StatusOpen,
		now,
	)
	if err != nil {
		return inbound.InitializePeriodOutput{}, fmt.Errorf("initialize period: %w", err)
	}

	if err := u.periodRepo.Create(ctx, p); err != nil {
		return inbound.InitializePeriodOutput{}, fmt.Errorf("initialize period: %w", err)
	}

	// 4. Adopt orphan expenses: link existing expenses in this date range to the new period.
	// This is critical for legacy data rollover.
	if err := u.expenseRepo.AdoptOrphanExpenses(ctx, input.HouseholdID, string(p.ID()), start, end); err != nil {
		// Log and continue — we don't want to block period creation if this fails 
		// (e.g. if column doesn't exist yet)
		fmt.Printf("Warning: failed to adopt orphan expenses: %v\n", err)
	}

	return inbound.InitializePeriodOutput{PeriodID: string(p.ID())}, nil
}

var _ inbound.InitializePeriodUseCase = InitializePeriodUseCase{}
