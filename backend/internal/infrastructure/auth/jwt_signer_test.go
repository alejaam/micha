package auth_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	infraauth "micha/backend/internal/infrastructure/auth"
)

// secret32 is a test secret exactly 32 bytes long.
const secret32 = "test-secret-that-is-32-bytes-lon"

// secret31 is one byte too short to exercise the guard.
const secret31 = "test-secret-31-bytes-long-short"

// --- NewJWTSigner -----------------------------------------------------------

func TestNewJWTSigner(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		secret  string
		wantErr error
	}{
		{name: "valid secret", secret: secret32, wantErr: nil},
		{name: "secret too short", secret: secret31, wantErr: infraauth.ErrWeakJWTSecret},
		{name: "empty secret", secret: "", wantErr: infraauth.ErrWeakJWTSecret},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := infraauth.NewJWTSigner(tc.secret)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewJWTSigner(%q): got %v; want %v", tc.secret, err, tc.wantErr)
			}
		})
	}
}

// --- JWTSigner.Sign ---------------------------------------------------------

func TestJWTSigner_Sign(t *testing.T) {
	t.Parallel()
	signer, err := infraauth.NewJWTSigner(secret32)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	cases := []struct {
		name    string
		userID  string
		email   string
		wantErr bool
	}{
		{name: "valid claims", userID: "user-1", email: "alice@example.com"},
		{name: "empty userID still signs", userID: "", email: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			token, err := signer.Sign(tc.userID, tc.email)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantErr && !strings.HasPrefix(token, "eyJ") {
				t.Errorf("token %q does not look like a JWT", token)
			}
		})
	}
}

// --- NewJWTValidator --------------------------------------------------------

func TestNewJWTValidator(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		secret  string
		wantErr error
	}{
		{name: "valid secret", secret: secret32, wantErr: nil},
		{name: "secret too short", secret: secret31, wantErr: infraauth.ErrWeakJWTSecret},
		{name: "empty secret", secret: "", wantErr: infraauth.ErrWeakJWTSecret},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := infraauth.NewJWTValidator(tc.secret)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewJWTValidator(%q): got %v; want %v", tc.secret, err, tc.wantErr)
			}
		})
	}
}

// --- JWTValidator.Validate --------------------------------------------------

func TestJWTValidator_Validate(t *testing.T) {
	t.Parallel()

	signer, err := infraauth.NewJWTSigner(secret32)
	if err != nil {
		t.Fatalf("setup signer: %v", err)
	}
	validator, err := infraauth.NewJWTValidator(secret32)
	if err != nil {
		t.Fatalf("setup validator: %v", err)
	}

	validToken, err := signer.Sign("user-1", "alice@example.com")
	if err != nil {
		t.Fatalf("setup sign: %v", err)
	}

	// Build an expired token manually.
	expiredToken := func() string {
		claims := jwt.MapClaims{
			"sub":   "user-2",
			"email": "bob@example.com",
			"exp":   time.Now().Add(-1 * time.Hour).Unix(),
			"iat":   time.Now().Add(-2 * time.Hour).Unix(),
		}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, _ := tok.SignedString([]byte(secret32))
		return signed
	}()

	// Build a token signed with a different secret.
	wrongSecretToken := func() string {
		signer2, _ := infraauth.NewJWTSigner("another-secret-that-is-32-bytes!")
		tok, _ := signer2.Sign("user-3", "carol@example.com")
		return tok
	}()

	cases := []struct {
		name        string
		token       string
		wantUserID  string
		wantEmail   string
		wantErrFrag string // non-empty means we expect an error containing this substring
	}{
		{
			name:       "valid token",
			token:      validToken,
			wantUserID: "user-1",
			wantEmail:  "alice@example.com",
		},
		{
			name:        "expired token",
			token:       expiredToken,
			wantErrFrag: "invalid token",
		},
		{
			name:        "wrong secret",
			token:       wrongSecretToken,
			wantErrFrag: "invalid token",
		},
		{
			name:        "garbage string",
			token:       "not.a.jwt",
			wantErrFrag: "invalid token",
		},
		{
			name:        "empty token",
			token:       "",
			wantErrFrag: "invalid token",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			userID, email, err := validator.Validate(tc.token)
			if tc.wantErrFrag != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrFrag)
				}
				if !strings.Contains(err.Error(), tc.wantErrFrag) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.wantErrFrag)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if userID != tc.wantUserID {
				t.Errorf("userID = %q; want %q", userID, tc.wantUserID)
			}
			if email != tc.wantEmail {
				t.Errorf("email = %q; want %q", email, tc.wantEmail)
			}
		})
	}
}
