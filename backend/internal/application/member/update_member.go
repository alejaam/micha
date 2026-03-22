package memberapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

var _ inbound.UpdateMemberUseCase = UpdateMemberUseCase{}

// UpdateMemberUseCase updates mutable member fields.
type UpdateMemberUseCase struct {
	repo outbound.MemberRepository
}

// NewUpdateMemberUseCase constructs UpdateMemberUseCase.
func NewUpdateMemberUseCase(repo outbound.MemberRepository) UpdateMemberUseCase {
	return UpdateMemberUseCase{repo: repo}
}

// Execute updates the member profile.
func (u UpdateMemberUseCase) Execute(ctx context.Context, input inbound.UpdateMemberInput) error {
	m, err := u.repo.FindByID(ctx, input.MemberID)
	if err != nil {
		return fmt.Errorf("update member: %w", err)
	}

	// Verify member belongs to the specified household.
	if m.HouseholdID() != input.HouseholdID {
		return fmt.Errorf("update member: member does not belong to household")
	}

	if err := m.UpdateProfile(input.Name, input.Email, input.MonthlySalaryCents); err != nil {
		return fmt.Errorf("update member: %w", err)
	}

	if err := u.repo.Update(ctx, m); err != nil {
		return fmt.Errorf("update member: %w", err)
	}

	slog.InfoContext(ctx, "update member", "member_id", input.MemberID)
	return nil
}
