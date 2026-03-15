package memberapp

import (
	"context"
	"fmt"
	"log/slog"
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
func (u RegisterMemberUseCase) Execute(ctx context.Context, input inbound.RegisterMemberInput) (inbound.RegisterMemberOutput, error) {
	m, err := member.NewFromAttributes(member.Attributes{
		ID:                 member.ID(u.idGenerator.NewID()),
		HouseholdID:        input.HouseholdID,
		Name:               input.Name,
		Email:              input.Email,
		MonthlySalaryCents: input.MonthlySalaryCents,
		UserID:             input.UserID,
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
