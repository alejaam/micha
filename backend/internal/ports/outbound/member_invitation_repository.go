package outbound

import (
	"context"

	"micha/backend/internal/domain/invitation"
)

// MemberInvitationRepository persists invitation codes for member onboarding.
type MemberInvitationRepository interface {
	Save(ctx context.Context, inv invitation.Invitation) error
}
