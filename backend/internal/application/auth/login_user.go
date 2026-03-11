package authapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// ErrInvalidCredentials is returned when the email or password is incorrect.
var ErrInvalidCredentials = errors.New("invalid credentials")

// TokenSigner signs a user ID and email into a JWT token.
type TokenSigner interface {
	Sign(userID, email string) (string, error)
}

// LoginUseCase authenticates a user and issues a JWT token.
type LoginUseCase struct {
	repo   outbound.UserRepository
	hasher PasswordHasher
	signer TokenSigner
}

// NewLoginUseCase constructs a LoginUseCase.
func NewLoginUseCase(repo outbound.UserRepository, hasher PasswordHasher, signer TokenSigner) LoginUseCase {
	return LoginUseCase{repo: repo, hasher: hasher, signer: signer}
}

// Execute looks up the user by email, verifies the password, and signs a token.
func (u LoginUseCase) Execute(ctx context.Context, input inbound.LoginInput) (inbound.LoginOutput, error) {
	foundUser, err := u.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return inbound.LoginOutput{}, ErrInvalidCredentials
		}
		return inbound.LoginOutput{}, fmt.Errorf("login: %w", err)
	}

	if err := u.hasher.Verify(input.Password, foundUser.PasswordHash()); err != nil {
		return inbound.LoginOutput{}, ErrInvalidCredentials
	}

	token, err := u.signer.Sign(foundUser.ID(), foundUser.Email())
	if err != nil {
		return inbound.LoginOutput{}, fmt.Errorf("login: %w", err)
	}

	slog.InfoContext(ctx, "user logged in", "user_id", foundUser.ID())
	return inbound.LoginOutput{Token: token}, nil
}

var _ inbound.LoginUseCase = LoginUseCase{}
