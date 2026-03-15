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
	// FindByUserIDGlobal returns any member record linked to the given user, regardless of household.
	// Returns shared.ErrNotFound when the user has no membership in any household.
	FindByUserIDGlobal(ctx context.Context, userID string) (member.Member, error)
	ListAllByHousehold(ctx context.Context, householdID string) ([]member.Member, error)
	ListByHousehold(ctx context.Context, householdID string, limit, offset int) ([]member.Member, error)
	// ListHouseholdIDsByUserID returns all household IDs that the user belongs to.
	ListHouseholdIDsByUserID(ctx context.Context, userID string) ([]string, error)
	Update(ctx context.Context, m member.Member) error
}
