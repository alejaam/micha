package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/invitation"
	"micha/backend/internal/ports/outbound"
)

// MemberInvitationRepository fulfils outbound.MemberInvitationRepository using PostgreSQL.
type MemberInvitationRepository struct {
	db *pgxpool.Pool
}

// NewMemberInvitationRepository constructs a MemberInvitationRepository backed by the given pool.
func NewMemberInvitationRepository(db *pgxpool.Pool) MemberInvitationRepository {
	return MemberInvitationRepository{db: db}
}

// Save persists an invitation code record.
func (r MemberInvitationRepository) Save(ctx context.Context, inv invitation.Invitation) error {
	attrs := inv.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO member_invitations (id, household_id, member_id, email, invite_code, expires_at, used_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		string(attrs.ID), attrs.HouseholdID, attrs.MemberID, attrs.Email, attrs.Code, attrs.ExpiresAt, attrs.UsedAt, attrs.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("member invitation repository save: %w", err)
	}
	return nil
}

var _ outbound.MemberInvitationRepository = MemberInvitationRepository{}
