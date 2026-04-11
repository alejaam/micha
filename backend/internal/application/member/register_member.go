package memberapp

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"sort"
	"strings"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/invitation"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// RegisterMemberUseCase creates a new member.
type RegisterMemberUseCase struct {
	repo         outbound.MemberRepository
	idGenerator  appshared.IDGenerator
	inviteRepo   outbound.MemberInvitationRepository
	inviteSender outbound.InviteCodeSender
	now          func() time.Time
	inviteTTL    time.Duration
}

// NewRegisterMemberUseCase constructs RegisterMemberUseCase.
func NewRegisterMemberUseCase(repo outbound.MemberRepository, idGenerator appshared.IDGenerator) RegisterMemberUseCase {
	return RegisterMemberUseCase{
		repo:        repo,
		idGenerator: idGenerator,
		now:         time.Now,
		inviteTTL:   24 * time.Hour,
	}
}

// NewRegisterMemberUseCaseWithInvites enables invitation code persistence + delivery for pending members.
func NewRegisterMemberUseCaseWithInvites(
	repo outbound.MemberRepository,
	idGenerator appshared.IDGenerator,
	inviteRepo outbound.MemberInvitationRepository,
	inviteSender outbound.InviteCodeSender,
) RegisterMemberUseCase {
	uc := NewRegisterMemberUseCase(repo, idGenerator)
	uc.inviteRepo = inviteRepo
	uc.inviteSender = inviteSender
	return uc
}

// Execute creates a member and stores it.
// If CallerEmail matches the new member's normalised email, the member is automatically linked to CallerUserID.
func (u RegisterMemberUseCase) Execute(ctx context.Context, input inbound.RegisterMemberInput) (inbound.RegisterMemberOutput, error) {
	if err := u.validateCreatePrivileges(ctx, input.HouseholdID, input.CallerMemberID); err != nil {
		return inbound.RegisterMemberOutput{}, fmt.Errorf("register member: %w", err)
	}

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

	// Auto-link: compare against the normalised email stored in the domain entity,
	// not the raw input string, to avoid mismatches caused by whitespace or casing.
	if m.UserID() == "" && input.CallerUserID != "" &&
		strings.EqualFold(m.Email(), strings.TrimSpace(input.CallerEmail)) {
		m.LinkUser(input.CallerUserID)
	}

	if err := u.repo.Save(ctx, m); err != nil {
		return inbound.RegisterMemberOutput{}, fmt.Errorf("register member: %w", err)
	}

	if err := u.createAndSendInviteIfPending(ctx, m); err != nil {
		return inbound.RegisterMemberOutput{}, fmt.Errorf("register member: %w", err)
	}

	slog.InfoContext(ctx, "register member", "member_id", string(m.ID()), "household_id", m.HouseholdID(), "user_linked", m.UserID() != "")
	return inbound.RegisterMemberOutput{MemberID: string(m.ID())}, nil
}

func (u RegisterMemberUseCase) validateCreatePrivileges(ctx context.Context, householdID, callerMemberID string) error {
	members, err := u.repo.ListAllByHousehold(ctx, householdID)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		return nil
	}

	ownerMemberID := findOwnerMemberID(members)
	if ownerMemberID == "" || ownerMemberID != callerMemberID {
		return shared.ErrForbidden
	}

	return nil
}

func (u RegisterMemberUseCase) createAndSendInviteIfPending(ctx context.Context, m member.Member) error {
	if !m.IsPending() || u.inviteRepo == nil || u.inviteSender == nil {
		return nil
	}

	code, err := generateInviteCode(6)
	if err != nil {
		return err
	}

	now := u.now()
	inv, err := invitation.NewFromAttributes(invitation.Attributes{
		ID:          invitation.ID(u.idGenerator.NewID()),
		HouseholdID: m.HouseholdID(),
		MemberID:    string(m.ID()),
		Email:       m.Email(),
		Code:        code,
		ExpiresAt:   now.Add(u.inviteTTL),
		CreatedAt:   now,
	})
	if err != nil {
		return err
	}

	if err := u.inviteRepo.Save(ctx, inv); err != nil {
		return err
	}

	return u.inviteSender.SendInviteCode(ctx, m.Email(), code)
}

func findOwnerMemberID(members []member.Member) string {
	linked := make([]member.Member, 0, len(members))
	for _, m := range members {
		if strings.TrimSpace(m.UserID()) != "" {
			linked = append(linked, m)
		}
	}
	if len(linked) == 0 {
		return ""
	}

	sort.SliceStable(linked, func(i, j int) bool {
		return linked[i].CreatedAt().Before(linked[j].CreatedAt())
	})

	return string(linked[0].ID())
}

func generateInviteCode(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[n.Int64()]
	}
	return string(b), nil
}

var _ inbound.RegisterMemberUseCase = RegisterMemberUseCase{}
