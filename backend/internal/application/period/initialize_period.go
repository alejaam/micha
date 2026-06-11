package periodapp

import (
	"context"
	"fmt"
	"time"

	appshared "micha/backend/internal/application/shared"
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
		now:           appshared.Now,
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

	// 3. Create initial period using shared helper.
	now := u.now()
	closingDay := h.Attributes().ClosingDay
	frequency := h.Attributes().PeriodFrequency

	result, err := CreateInitialPeriod(ctx, u.periodRepo, u.idGenerator, input.HouseholdID, closingDay, frequency, now)
	if err != nil {
		return inbound.InitializePeriodOutput{}, fmt.Errorf("initialize period: %w", err)
	}

	// 4. Adopt orphan expenses: link existing expenses in this date range to the new period.
	if err := u.expenseRepo.AdoptOrphanExpenses(ctx, input.HouseholdID, result.PeriodID, result.StartDate, result.EndDate); err != nil {
		fmt.Printf("Warning: failed to adopt orphan expenses: %v\n", err)
	}

	return inbound.InitializePeriodOutput{PeriodID: result.PeriodID}, nil
}

var _ inbound.InitializePeriodUseCase = InitializePeriodUseCase{}
