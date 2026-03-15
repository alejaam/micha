package memberapp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"micha/backend/internal/domain/member"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// IDGenerator abstracts id generation for testability.
type IDGenerator interface {
	NewID() string
}

// RegisterMemberUseCase creates a new member.
type RegisterMemberUseCase struct {
	repo        outbound.MemberRepository
	idGenerator IDGenerator
	now         func() time.Time
}

// NewRegisterMemberUseCase constructs RegisterMemberUseCase.
func NewRegisterMemberUseCase(repo outbound.MemberRepository, idGenerator IDGenerator) RegisterMemberUseCase {
	return RegisterMemberUseCase{
		repo:        repo,
		idGenerator: idGenerator,
		now:         time.Now,
	}
}

// Execute creates a member and stores it.
// If CallerEmail matches the new member's email, the member is automatically linked to CallerUserID.
func (u RegisterMemberUseCase) Execute(ctx context.Context, input inbound.RegisterMemberInput) (inbound.RegisterMemberOutput, error) {
	// Auto-link: if the caller's email matches the member email, set UserID.
	linkedUserID := input.UserID
	if linkedUserID == "" && input.CallerUserID != "" &&
		strings.EqualFold(strings.TrimSpace(input.CallerEmail), strings.TrimSpace(input.Email)) {
		linkedUserID = input.CallerUserID
	}

	m, err := member.NewFromAttributes(member.Attributes{
		ID:                 member.ID(u.idGenerator.NewID()),
		HouseholdID:        input.HouseholdID,
		Name:               input.Name,
		Email:              input.Email,
		MonthlySalaryCents: input.MonthlySalaryCents,
		UserID:             linkedUserID,
		CreatedAt:          u.now(),
	})
	if err != nil {
		return inbound.RegisterMemberOutput{}, fmt.Errorf("register member: %w", err)
	}

	if err := u.repo.Save(ctx, m); err != nil {
		return inbound.RegisterMemberOutput{}, fmt.Errorf("register member: %w", err)
	}

	slog.InfoContext(ctx, "register member", "member_id", string(m.ID()), "household_id", m.HouseholdID(), "user_linked", m.UserID() != "")
	return inbound.RegisterMemberOutput{MemberID: string(m.ID())}, nil
}

var _ inbound.RegisterMemberUseCase = RegisterMemberUseCase{}
