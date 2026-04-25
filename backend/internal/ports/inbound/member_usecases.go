package inbound

import (
	"context"

	"micha/backend/internal/domain/member"
)

// RegisterMemberInput contains required data to register a member.
type RegisterMemberInput struct {
	HouseholdID        string
	Name               string
	Email              string
	MonthlySalaryCents int64
	// CallerMemberID is the authenticated member id in the target household.
	// It may be empty only in bootstrap flow (first member creation).
	CallerMemberID string
	// UserID optionally links the new member to an authenticated user. Empty means no link.
	UserID string
	// CallerUserID is the authenticated user's ID — used for auto-linking when emails match.
	CallerUserID string
	// CallerEmail is the authenticated user's email — used for auto-linking when emails match.
	CallerEmail string
}

// RegisterMemberOutput contains created member identifiers.
type RegisterMemberOutput struct {
	MemberID string
}

// ListMembersQuery holds listing parameters for members by household.
type ListMembersQuery struct {
	HouseholdID string
	Limit       int
	Offset      int
}

// UpdateMemberInput contains mutable fields for updating a member.
type UpdateMemberInput struct {
	MemberID           string
	HouseholdID        string
	Name               string
	Email              string
	MonthlySalaryCents int64
}

// DeleteMemberInput contains data required to delete a member.
type DeleteMemberInput struct {
	MemberID    string
	HouseholdID string
}

// CalculateRemainingSalaryInput defines one period request for member finance reporting.
type CalculateRemainingSalaryInput struct {
	HouseholdID string
	MemberID    string
	Year        int
	Month       int
}

// CalculateRemainingSalaryOutput contains member finance totals for one period.
type CalculateRemainingSalaryOutput struct {
	HouseholdID           string
	MemberID              string
	Year                  int
	Month                 int
	MonthlySalaryCents    int64
	PersonalExpensesCents int64
	AllocatedDebtCents    int64
	RemainingSalaryCents  int64
}

type RegisterMemberUseCase interface {
	Execute(ctx context.Context, input RegisterMemberInput) (RegisterMemberOutput, error)
}

type ListMembersUseCase interface {
	Execute(ctx context.Context, query ListMembersQuery) ([]member.Member, error)
}

type UpdateMemberUseCase interface {
	Execute(ctx context.Context, input UpdateMemberInput) error
}

type DeleteMemberUseCase interface {
	Execute(ctx context.Context, input DeleteMemberInput) error
}

type CalculateRemainingSalaryUseCase interface {
	Execute(ctx context.Context, input CalculateRemainingSalaryInput) (CalculateRemainingSalaryOutput, error)
}
