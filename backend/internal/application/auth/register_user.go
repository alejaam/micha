package authapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// IDGenerator generates unique string identifiers.
type IDGenerator interface {
	NewID() string
}

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// RegisterUserUseCase creates new user accounts.
type RegisterUserUseCase struct {
	repo   outbound.UserRepository
	idGen  IDGenerator
	hasher PasswordHasher
}

// NewRegisterUserUseCase constructs a RegisterUserUseCase.
func NewRegisterUserUseCase(repo outbound.UserRepository, idGen IDGenerator, hasher PasswordHasher) RegisterUserUseCase {
	return RegisterUserUseCase{repo: repo, idGen: idGen, hasher: hasher}
}

// Execute validates, hashes the password, and persists a new user.
func (u RegisterUserUseCase) Execute(ctx context.Context, input inbound.RegisterUserInput) (inbound.RegisterUserOutput, error) {
	if input.Email == "" {
		return inbound.RegisterUserOutput{}, fmt.Errorf("register user: email is required")
	}

	if input.Password == "" {
		return inbound.RegisterUserOutput{}, fmt.Errorf("register user: password is required")
	}

	hash, err := u.hasher.Hash(input.Password)
	if err != nil {
		return inbound.RegisterUserOutput{}, fmt.Errorf("register user: %w", err)
	}

	now := time.Now()
	newUser, err := user.New(u.idGen.NewID(), input.Email, hash, now)
	if err != nil {
		return inbound.RegisterUserOutput{}, fmt.Errorf("register user: %w", err)
	}

	if err := u.repo.Save(ctx, newUser); err != nil {
		return inbound.RegisterUserOutput{}, fmt.Errorf("register user: %w", err)
	}

	slog.InfoContext(ctx, "user registered", "user_id", newUser.ID())
	return inbound.RegisterUserOutput{UserID: newUser.ID()}, nil
}

var _ inbound.RegisterUserUseCase = RegisterUserUseCase{}
