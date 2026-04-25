package periodapp

import (
	"context"
	"fmt"
	"time"

	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type TransitionToReviewUseCase struct {
	periodRepo outbound.PeriodRepository
	memberRepo outbound.MemberRepository
	now        func() time.Time
}

func NewTransitionToReviewUseCase(
	periodRepo outbound.PeriodRepository,
	memberRepo outbound.MemberRepository,
) TransitionToReviewUseCase {
	return TransitionToReviewUseCase{
		periodRepo: periodRepo,
		memberRepo: memberRepo,
		now:        time.Now,
	}
}

func (u TransitionToReviewUseCase) Execute(ctx context.Context, input inbound.TransitionToReviewInput) (inbound.TransitionToReviewOutput, error) {
	// 1. Validate that the member exists and belongs to the household.
	m, err := u.memberRepo.FindByUserID(ctx, input.HouseholdID, input.CurrentUserID)
	if err != nil {
		return inbound.TransitionToReviewOutput{}, fmt.Errorf("transition to review: %w", err)
	}
	if m.HouseholdID() != input.HouseholdID {
		return inbound.TransitionToReviewOutput{}, fmt.Errorf("transition to review: %w", shared.ErrForbidden)
	}

	// 2. Retrieve the period and validate its current status.
	p, err := u.periodRepo.GetByID(ctx, period.ID(input.PeriodID))
	if err != nil {
		return inbound.TransitionToReviewOutput{}, fmt.Errorf("transition to review: %w", err)
	}

	if p.HouseholdID() != input.HouseholdID {
		return inbound.TransitionToReviewOutput{}, fmt.Errorf("transition to review: %w", shared.ErrForbidden)
	}

	if p.Status() != period.StatusOpen {
		return inbound.TransitionToReviewOutput{}, fmt.Errorf("transition to review: only open periods can be moved to review")
	}

	// 3. Create the updated period in domain.
	attrs := p.Attributes()
	attrs.Status = period.StatusReview
	attrs.UpdatedAt = u.now()

	updatedPeriod, err := period.NewFromAttributes(attrs)
	if err != nil {
		return inbound.TransitionToReviewOutput{}, fmt.Errorf("transition to review: %w", err)
	}

	// 4. Persist the change.
	if err := u.periodRepo.Update(ctx, updatedPeriod); err != nil {
		return inbound.TransitionToReviewOutput{}, fmt.Errorf("transition to review: %w", err)
	}

	return inbound.TransitionToReviewOutput{Success: true}, nil
}

var _ inbound.TransitionToReviewUseCase = TransitionToReviewUseCase{}
