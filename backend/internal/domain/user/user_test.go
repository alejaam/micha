package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"micha/backend/internal/domain/user"
)

var baseAttrs = user.UserAttributes{
	ID:           "user-1",
	Email:        "alice@example.com",
	PasswordHash: "$2a$12$somehashvalue",
	CreatedAt:    time.Now(),
}

func TestNew_ValidUser(t *testing.T) {
	t.Parallel()

	u, err := user.New(baseAttrs.ID, baseAttrs.Email, baseAttrs.PasswordHash, baseAttrs.CreatedAt)
	require.NoError(t, err)
	require.Equal(t, baseAttrs.ID, u.ID())
	require.Equal(t, baseAttrs.Email, u.Email())
	require.Equal(t, baseAttrs.PasswordHash, u.PasswordHash())
}

func TestNew_InvalidEmail(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"missing at sign", "invalidemail"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := user.New("id-1", tc.email, "somehash", time.Now())
			require.ErrorIs(t, err, user.ErrInvalidEmail)
		})
	}
}

func TestNew_WeakPassword(t *testing.T) {
	t.Parallel()

	_, err := user.New("id-1", "alice@example.com", "", time.Now())
	require.ErrorIs(t, err, user.ErrWeakPassword)
}

func TestNewFromAttributes_ValidUser(t *testing.T) {
	t.Parallel()

	u, err := user.NewFromAttributes(baseAttrs)
	require.NoError(t, err)

	attrs := u.Attributes()
	require.Equal(t, baseAttrs.ID, attrs.ID)
	require.Equal(t, baseAttrs.Email, attrs.Email)
	require.Equal(t, baseAttrs.PasswordHash, attrs.PasswordHash)
}

func TestNewFromAttributes_InvalidEmail(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"no at sign", "notanemail"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			attrs := baseAttrs
			attrs.Email = tc.email
			_, err := user.NewFromAttributes(attrs)
			require.ErrorIs(t, err, user.ErrInvalidEmail)
		})
	}
}

func TestNewFromAttributes_WeakPassword(t *testing.T) {
	t.Parallel()

	attrs := baseAttrs
	attrs.PasswordHash = "   "
	_, err := user.NewFromAttributes(attrs)
	require.ErrorIs(t, err, user.ErrWeakPassword)
}
