package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"micha/backend/internal/ports/outbound"
)

// InviteDeliveryDevLog is the explicit development-only delivery mode.
const InviteDeliveryDevLog = "dev-log"

// ErrInviteDeliveryNotConfigured is returned when no invite delivery adapter is configured.
var ErrInviteDeliveryNotConfigured = errors.New("invite delivery is not configured")

// NewInviteSender selects the invite delivery adapter from explicit configuration.
// An empty mode yields a sender that fails explicitly on every send.
func NewInviteSender(mode, environment string) (outbound.InviteCodeSender, error) {
	switch strings.TrimSpace(mode) {
	case "":
		return NotConfiguredInviteSender{}, nil
	case InviteDeliveryDevLog:
		if strings.EqualFold(strings.TrimSpace(environment), "production") {
			return nil, fmt.Errorf("invite delivery mode %q is not allowed in production", InviteDeliveryDevLog)
		}
		return DevLogInviteSender{}, nil
	default:
		return nil, fmt.Errorf("unknown invite delivery mode %q", mode)
	}
}

// NotConfiguredInviteSender always fails: no delivery adapter has been configured.
type NotConfiguredInviteSender struct{}

func (NotConfiguredInviteSender) SendInviteCode(_ context.Context, _, _ string) error {
	return ErrInviteDeliveryNotConfigured
}

// DevLogInviteSender logs a masked trace instead of delivering. Development only.
type DevLogInviteSender struct{}

func (DevLogInviteSender) SendInviteCode(_ context.Context, email, code string) error {
	slog.Info("dev invite delivery (not sent)", "email", maskEmail(email), "code", maskCode(code))
	return nil
}

func maskCode(code string) string {
	if len(code) <= 2 {
		return strings.Repeat("*", len(code))
	}
	return strings.Repeat("*", len(code)-2) + code[len(code)-2:]
}

func maskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return "***"
	}
	return email[:1] + "***" + email[at:]
}

var (
	_ outbound.InviteCodeSender = NotConfiguredInviteSender{}
	_ outbound.InviteCodeSender = DevLogInviteSender{}
)
