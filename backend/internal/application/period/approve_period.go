package periodapp

import (
	"context"
	"fmt"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/period"
	periodapproval "micha/backend/internal/domain/period_approval"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type ApprovePeriodUseCase struct {
	approvalRepo outbound.PeriodApprovalRepository
	periodRepo   outbound.PeriodRepository
	memberRepo   outbound.MemberRepository
	idGenerator  appshared.IDGenerator
	now          func() time.Time
}

func NewApprovePeriodUseCase(
	approvalRepo outbound.PeriodApprovalRepository,
	periodRepo outbound.PeriodRepository,
	memberRepo outbound.MemberRepository,
	idGenerator appshared.IDGenerator,
) ApprovePeriodUseCase {
	return ApprovePeriodUseCase{
		approvalRepo: approvalRepo,
		periodRepo:   periodRepo,
		memberRepo:   memberRepo,
		idGenerator:  idGenerator,
		now:          time.Now,
	}
}

func (u ApprovePeriodUseCase) Execute(ctx context.Context, input inbound.ApprovePeriodInput) (inbound.ApprovePeriodOutput, error) {
	// 1. Validate that the member exists and belongs to the household.
	m, err := u.memberRepo.FindByUserID(ctx, input.HouseholdID, input.CurrentUserID)
	if err != nil {
		return inbound.ApprovePeriodOutput{}, fmt.Errorf("approve period: %w", err)
	}
	if m.HouseholdID() != input.HouseholdID {
		return inbound.ApprovePeriodOutput{}, fmt.Errorf("approve period: %w", shared.ErrForbidden)
	}

	// 2. Retrieve the period and validate it's in review status.
	p, err := u.periodRepo.GetByID(ctx, period.ID(input.PeriodID))
	if err != nil {
		return inbound.ApprovePeriodOutput{}, fmt.Errorf("approve period: %w", err)
	}

	if p.HouseholdID() != input.HouseholdID {
		return inbound.ApprovePeriodOutput{}, fmt.Errorf("approve period: %w", shared.ErrForbidden)
	}

	if p.Status() != period.StatusReview {
		return inbound.ApprovePeriodOutput{}, fmt.Errorf("approve period: only periods in review status can be approved/objected")
	}

	// 3. Check if an approval already exists for this member/period.
	existing, err := u.approvalRepo.GetByMemberAndPeriod(ctx, string(m.ID()), input.PeriodID)
	
	now := u.now()
	var a periodapproval.PeriodApproval
	
	if err == nil {
		// Update existing
		attrs := existing.Attributes()
		attrs.Status = periodapproval.ApprovalStatus(input.Status)
		attrs.Comment = input.Comment
		attrs.UpdatedAt = now
		a, err = periodapproval.NewFromAttributes(attrs)
	} else {
		// Create new
		a, err = periodapproval.New(
			periodapproval.ID(u.idGenerator.NewID()),
			string(m.ID()),
			input.PeriodID,
			periodapproval.ApprovalStatus(input.Status),
			input.Comment,
			now,
		)
	}

	if err != nil {
		return inbound.ApprovePeriodOutput{}, fmt.Errorf("approve period: %w", err)
	}

	// 4. Persist
	if err := u.approvalRepo.Save(ctx, a); err != nil {
		return inbound.ApprovePeriodOutput{}, fmt.Errorf("approve period: %w", err)
	}

	return inbound.ApprovePeriodOutput{ApprovalID: string(a.ID())}, nil
}

var _ inbound.ApprovePeriodUseCase = ApprovePeriodUseCase{}
