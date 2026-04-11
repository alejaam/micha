package outbound

import "context"

// InviteCodeSender sends invitation codes to member emails.
type InviteCodeSender interface {
	SendInviteCode(ctx context.Context, email, code string) error
}
