package inbound

import "context"

// TransitionToReviewInput defines the required data to move a period to review status.
type TransitionToReviewInput struct {
	HouseholdID   string
	PeriodID      string
	CurrentUserID string
}

type TransitionToReviewOutput struct {
	Success bool
}

// TransitionToReviewUseCase contract.
type TransitionToReviewUseCase interface {
	Execute(ctx context.Context, input TransitionToReviewInput) (TransitionToReviewOutput, error)
}

// ApprovePeriodInput defines the required data to approve or object a period.
type ApprovePeriodInput struct {
	HouseholdID   string
	PeriodID      string
	CurrentUserID string
	Status        string // "approved" | "objected"
	Comment       string
}

type ApprovePeriodOutput struct {
	ApprovalID string
}

// ApprovePeriodUseCase contract.
type ApprovePeriodUseCase interface {
	Execute(ctx context.Context, input ApprovePeriodInput) (ApprovePeriodOutput, error)
}

// ClosePeriodInput defines the required data to close a period and trigger rollover.
type ClosePeriodInput struct {
	HouseholdID   string
	PeriodID      string
	CurrentUserID string
	Force         bool // If true, owner can close even with objections
}

type ClosePeriodOutput struct {
	NextPeriodID string
}

// ClosePeriodUseCase contract.
type ClosePeriodUseCase interface {
	Execute(ctx context.Context, input ClosePeriodInput) (ClosePeriodOutput, error)
}
