package memberapp

import (
	"context"
	"fmt"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
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
		limit = appshared.DefaultLimit
	}
	if limit > appshared.MaxLimit {
		limit = appshared.MaxLimit
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

var _ inbound.ListMembersUseCase = ListMembersUseCase{}
