package authapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authapp "micha/backend/internal/application/auth"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/inbound"
)

// --- mock helpers -----------------------------------------------------------

type staticIDGen string

func (s staticIDGen) NewID() string { return string(s) }

// mockUserRepo implements outbound.UserRepository using in-memory storage.
type mockUserRepo struct {
	byEmail map[string]user.User
	saveErr error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{byEmail: make(map[string]user.User)}
}

func (m *mockUserRepo) Save(_ context.Context, u user.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.byEmail[u.Email()] = u
	return nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (user.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return user.User{}, shared.ErrNotFound
	}
	return u, nil
}

// mockHasher implements outbound.PasswordHasher.
type mockHasher struct {
	hashErr   error
	verifyErr error
}

func (h *mockHasher) Hash(password string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return "hashed:" + password, nil
}

func (h *mockHasher) Verify(password, hash string) error {
	if h.verifyErr != nil {
		return h.verifyErr
	}
	if "hashed:"+password != hash {
		return errors.New("wrong password")
	}
	return nil
}

// mockSigner implements outbound.TokenSigner.
type mockSigner struct {
	signErr error
}

func (s *mockSigner) Sign(userID, _ string) (string, error) {
	if s.signErr != nil {
		return "", s.signErr
	}
	return "token:" + userID, nil
}

// seedUser saves a user directly into the mock repo, bypassing the use case.
func seedUser(t *testing.T, repo *mockUserRepo, id, email, password string) {
	t.Helper()
	h := &mockHasher{}
	hash, _ := h.Hash(password)
	u, err := user.NewFromAttributes(user.UserAttributes{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("seedUser: %v", err)
	}
	repo.byEmail[email] = u
}

// --- RegisterUserUseCase tests ----------------------------------------------

func TestRegisterUser_Success(t *testing.T) {
	t.Parallel()
	repo := newMockUserRepo()
	uc := authapp.NewRegisterUserUseCase(repo, staticIDGen("user-1"), &mockHasher{})

	out, err := uc.Execute(context.Background(), inbound.RegisterUserInput{
		Email:    "ale@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.UserID != "user-1" {
		t.Errorf("UserID = %q; want %q", out.UserID, "user-1")
	}
	if _, findErr := repo.FindByEmail(context.Background(), "ale@example.com"); findErr != nil {
		t.Error("user was not persisted in the repo")
	}
}

func TestRegisterUser_TableDriven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   inbound.RegisterUserInput
		hashErr error
		saveErr error
		wantErr bool
	}{
		{
			name:    "empty email",
			input:   inbound.RegisterUserInput{Email: "", Password: "secret"},
			wantErr: true,
		},
		{
			name:    "empty password",
			input:   inbound.RegisterUserInput{Email: "ale@example.com", Password: ""},
			wantErr: true,
		},
		{
			name:    "hasher error",
			input:   inbound.RegisterUserInput{Email: "ale@example.com", Password: "secret"},
			hashErr: errors.New("bcrypt failed"),
			wantErr: true,
		},
		{
			name:    "repo save error",
			input:   inbound.RegisterUserInput{Email: "ale@example.com", Password: "secret"},
			saveErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := newMockUserRepo()
			repo.saveErr = tc.saveErr
			hasher := &mockHasher{hashErr: tc.hashErr}
			uc := authapp.NewRegisterUserUseCase(repo, staticIDGen("u-1"), hasher)

			_, err := uc.Execute(context.Background(), tc.input)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// --- LoginUseCase tests -----------------------------------------------------

func TestLogin_Success(t *testing.T) {
	t.Parallel()
	repo := newMockUserRepo()
	seedUser(t, repo, "user-1", "ale@example.com", "secret123")

	uc := authapp.NewLoginUseCase(repo, &mockHasher{}, &mockSigner{})

	out, err := uc.Execute(context.Background(), inbound.LoginInput{
		Email:    "ale@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Token != "token:user-1" {
		t.Errorf("Token = %q; want %q", out.Token, "token:user-1")
	}
}

func TestLogin_TableDriven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		email     string
		password  string
		verifyErr error
		signErr   error
		wantErr   error
	}{
		{
			name:     "user not found",
			email:    "nobody@example.com",
			password: "secret",
			wantErr:  shared.ErrInvalidCredentials,
		},
		{
			name:      "wrong password",
			email:     "ale@example.com",
			password:  "wrongpassword",
			verifyErr: errors.New("mismatch"),
			wantErr:   shared.ErrInvalidCredentials,
		},
		{
			name:     "signer error propagates",
			email:    "ale@example.com",
			password: "secret123",
			signErr:  errors.New("signing failed"),
			wantErr:  nil, // not ErrInvalidCredentials — a wrapped generic error
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := newMockUserRepo()
			seedUser(t, repo, "user-1", "ale@example.com", "secret123")

			hasher := &mockHasher{verifyErr: tc.verifyErr}
			signer := &mockSigner{signErr: tc.signErr}
			uc := authapp.NewLoginUseCase(repo, hasher, signer)

			_, err := uc.Execute(context.Background(), inbound.LoginInput{
				Email:    tc.email,
				Password: tc.password,
			})

			switch {
			case tc.name == "signer error propagates":
				// Should return a non-nil error that is NOT ErrInvalidCredentials.
				if err == nil {
					t.Error("expected error, got nil")
				}
				if errors.Is(err, shared.ErrInvalidCredentials) {
					t.Errorf("signer error should not map to ErrInvalidCredentials, got %v", err)
				}
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("want %v, got %v", tc.wantErr, err)
				}
			default:
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
