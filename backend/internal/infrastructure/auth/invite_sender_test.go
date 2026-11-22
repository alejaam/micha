package auth

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestNewInviteSender_NoConfigurationFailsExplicitly(t *testing.T) {
	sender, err := NewInviteSender("", "development")
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	if sendErr := sender.SendInviteCode(context.Background(), "a@b.com", "123456"); !errors.Is(sendErr, ErrInviteDeliveryNotConfigured) {
		t.Fatalf("expected ErrInviteDeliveryNotConfigured, got %v", sendErr)
	}
}

func TestNewInviteSender_UnknownModeIsError(t *testing.T) {
	if _, err := NewInviteSender("carrier-pigeon", "development"); err == nil {
		t.Fatal("expected error for unknown delivery mode")
	}
}

func TestNewInviteSender_DevLogRefusedInProduction(t *testing.T) {
	if _, err := NewInviteSender(InviteDeliveryDevLog, "production"); err == nil {
		t.Fatal("expected dev-log to be refused in production")
	}
}

func TestDevLogInviteSender_DoesNotLogFullCodeOrEmail(t *testing.T) {
	buf := captureLogs(t)
	sender, err := NewInviteSender(InviteDeliveryDevLog, "development")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sendErr := sender.SendInviteCode(context.Background(), "alice@example.com", "123456"); sendErr != nil {
		t.Fatalf("unexpected send error: %v", sendErr)
	}
	out := buf.String()
	if strings.Contains(out, "123456") {
		t.Fatalf("full code leaked in logs: %s", out)
	}
	if strings.Contains(out, "alice@example.com") {
		t.Fatalf("full email leaked in logs: %s", out)
	}
	if !strings.Contains(out, "56") {
		t.Fatalf("expected masked code hint in logs: %s", out)
	}
}

func TestNotConfiguredSender_DoesNotLogSecrets(t *testing.T) {
	buf := captureLogs(t)
	sender, _ := NewInviteSender("", "development")
	_ = sender.SendInviteCode(context.Background(), "alice@example.com", "123456")
	if out := buf.String(); strings.Contains(out, "123456") || strings.Contains(out, "alice@example.com") {
		t.Fatalf("secrets leaked in logs: %s", out)
	}
}
