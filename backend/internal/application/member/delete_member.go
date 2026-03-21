package memberapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/member"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ErrLastMember is returned when attempting to delete the last member of a household.
var ErrLastMember = errors.New("cannot delete the last member of a household")

var _ inbound.DeleteMemberUseCase = DeleteMemberUseCase{}

// DeleteMemberUseCase soft-deletes a member.
type DeleteMemberUseCase struct {
	repo outbound.MemberRepository
}

// NewDeleteMemberUseCase constructs DeleteMemberUseCase.
func NewDeleteMemberUseCase(repo outbound.MemberRepository) DeleteMemberUseCase {
	return DeleteMemberUseCase{repo: repo}
}

// Execute soft-deletes the member after validating business rules.
func (u DeleteMemberUseCase) Execute(ctx context.Context, input inbound.DeleteMemberInput) error {
	m, err := u.repo.FindByID(ctx, input.MemberID)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	// Verify member belongs to the specified household.
	if m.HouseholdID() != input.HouseholdID {
		return fmt.Errorf("delete member: %w", member.ErrInvalidName) // reuse domain error for "not found in this household"
	}

	// Prevent deleting the last member.
	count, err := u.repo.CountActiveByHousehold(ctx, input.HouseholdID)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}
	if count <= 1 {
		return ErrLastMember
	}

	if err := u.repo.Delete(ctx, input.MemberID); err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	slog.InfoContext(ctx, "delete member", "member_id", input.MemberID)
	return nil
}
