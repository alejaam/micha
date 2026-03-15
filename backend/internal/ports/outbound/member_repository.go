package outbound

import (
	"context"

	"micha/backend/internal/domain/member"
)

// MemberRepository defines persistence operations required by member use cases.
type MemberRepository interface {
	Save(ctx context.Context, m member.Member) error
	FindByID(ctx context.Context, id string) (member.Member, error)
	// FindByUserID returns the member linked to the given user within a household.
	// Returns shared.ErrNotFound when no link exists.
	FindByUserID(ctx context.Context, householdID, userID string) (member.Member, error)
	ListAllByHousehold(ctx context.Context, householdID string) ([]member.Member, error)
	ListByHousehold(ctx context.Context, householdID string, limit, offset int) ([]member.Member, error)
	Update(ctx context.Context, m member.Member) error
}
