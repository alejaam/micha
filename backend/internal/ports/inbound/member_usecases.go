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

type RegisterMemberUseCase interface {
	Execute(ctx context.Context, input RegisterMemberInput) (RegisterMemberOutput, error)
}

type ListMembersUseCase interface {
	Execute(ctx context.Context, query ListMembersQuery) ([]member.Member, error)
}
