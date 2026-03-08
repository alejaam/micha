package memberapp

import (
	"context"
	"fmt"

	"micha/backend/internal/domain/member"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ListMembersUseCase lists members for a household with pagination.
type ListMembersUseCase struct {
	repo outbound.MemberRepository
}

// NewListMembersUseCase constructs ListMembersUseCase.
func NewListMembersUseCase(repo outbound.MemberRepository) ListMembersUseCase {
	return ListMembersUseCase{repo: repo}
}

// Execute lists members for the given household.
func (u ListMembersUseCase) Execute(ctx context.Context, query inbound.ListMembersQuery) ([]member.Member, error) {
	if query.HouseholdID == "" {
		return nil, fmt.Errorf("list members: household_id is required")
	}

	limit := query.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	members, err := u.repo.ListByHousehold(ctx, query.HouseholdID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	return members, nil
}
