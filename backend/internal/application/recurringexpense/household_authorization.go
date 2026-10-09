package recurringexpenseapp

import (
	"context"
	"errors"
	"fmt"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

func authorizeActiveHouseholdMember(ctx context.Context, memberRepo outbound.MemberRepository, householdID, currentUserID string) error {
	if _, err := memberRepo.FindByUserID(ctx, householdID, currentUserID); err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return shared.ErrForbidden
		}
		return fmt.Errorf("authorize active household member: %w", err)
	}
	return nil
}
