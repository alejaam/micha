package periodapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/period"
	periodapproval "micha/backend/internal/domain/period_approval"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// GetPeriodConsensusUseCase returns the approval consensus summary for a period.
type GetPeriodConsensusUseCase struct {
	periodRepo   outbound.PeriodRepository
	approvalRepo outbound.PeriodApprovalRepository
	memberRepo   outbound.MemberRepository
}

// NewGetPeriodConsensusUseCase constructs the use case.
func NewGetPeriodConsensusUseCase(
	periodRepo outbound.PeriodRepository,
	approvalRepo outbound.PeriodApprovalRepository,
	memberRepo outbound.MemberRepository,
) GetPeriodConsensusUseCase {
	return GetPeriodConsensusUseCase{
		periodRepo:   periodRepo,
		approvalRepo: approvalRepo,
		memberRepo:   memberRepo,
	}
}

// Execute returns the number of approved members, total active members, and approval percentage.
func (u GetPeriodConsensusUseCase) Execute(ctx context.Context, input inbound.GetPeriodConsensusInput) (inbound.GetPeriodConsensusOutput, error) {
	// Verify the period exists and belongs to the household.
	p, err := u.periodRepo.GetByID(ctx, period.ID(input.PeriodID))
	if err != nil {
		return inbound.GetPeriodConsensusOutput{}, fmt.Errorf("get consensus: %w", err)
	}
	if p.HouseholdID() != input.HouseholdID {
		return inbound.GetPeriodConsensusOutput{}, fmt.Errorf("get consensus: period not found in household")
	}

	// Count total active members.
	total, err := u.memberRepo.CountActiveByHousehold(ctx, input.HouseholdID)
	if err != nil {
		return inbound.GetPeriodConsensusOutput{}, fmt.Errorf("get consensus: %w", err)
	}

	// Get all approvals for this period.
	approvals, err := u.approvalRepo.ListByPeriod(ctx, input.PeriodID)
	if err != nil {
		return inbound.GetPeriodConsensusOutput{}, fmt.Errorf("get consensus: %w", err)
	}

	approved := 0
	for _, a := range approvals {
		if a.Status() == periodapproval.ApprovalStatusApproved {
			approved++
		}
	}

	percent := 0.0
	if total > 0 {
		percent = float64(approved) / float64(total) * 100.0
	}

	return inbound.GetPeriodConsensusOutput{
		Approved: approved,
		Total:    total,
		Percent:  percent,
	}, nil
}

var _ inbound.GetPeriodConsensusUseCase = GetPeriodConsensusUseCase{}
