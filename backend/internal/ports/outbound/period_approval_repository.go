package outbound

import (
	"context"
	periodapproval "micha/backend/internal/domain/period_approval"
)

// PeriodApprovalRepository defines the contract for persisting and retrieving period approvals.
type PeriodApprovalRepository interface {
	// Save persists a new or existing period approval.
	Save(ctx context.Context, a periodapproval.PeriodApproval) error

	// GetByMemberAndPeriod retrieves the approval for a specific member in a specific period.
	GetByMemberAndPeriod(ctx context.Context, memberID, periodID string) (periodapproval.PeriodApproval, error)

	// ListByPeriod retrieves all approvals for a specific period.
	ListByPeriod(ctx context.Context, periodID string) ([]periodapproval.PeriodApproval, error)

	// DeleteAllByPeriod removes all approvals for a period (useful if transitioning back to open).
	DeleteAllByPeriod(ctx context.Context, periodID string) error
}
