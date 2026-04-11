package auth

import (
	"context"
	"log/slog"

	"micha/backend/internal/ports/outbound"
)

// LogInviteCodeSender is a temporary sender implementation used while email delivery is not wired.
type LogInviteCodeSender struct{}

func NewLogInviteCodeSender() LogInviteCodeSender { return LogInviteCodeSender{} }

func (s LogInviteCodeSender) SendInviteCode(_ context.Context, email, code string) error {
	slog.Info("invite code generated", "email", email, "code", code)
	return nil
}

var _ outbound.InviteCodeSender = LogInviteCodeSender{}
