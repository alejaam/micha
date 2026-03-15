package authapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// RegisterUserUseCase creates new user accounts.
type RegisterUserUseCase struct {
	repo   outbound.UserRepository
	idGen  appshared.IDGenerator
	hasher outbound.PasswordHasher
	now    func() time.Time
}

// NewRegisterUserUseCase constructs a RegisterUserUseCase.
func NewRegisterUserUseCase(repo outbound.UserRepository, idGen appshared.IDGenerator, hasher outbound.PasswordHasher) RegisterUserUseCase {
	return RegisterUserUseCase{repo: repo, idGen: idGen, hasher: hasher, now: time.Now}
}

// Execute validates, hashes the password, and persists a new user.
func (u RegisterUserUseCase) Execute(ctx context.Context, input inbound.RegisterUserInput) (inbound.RegisterUserOutput, error) {
	if input.Password == "" {
		return inbound.RegisterUserOutput{}, fmt.Errorf("register user: %w", user.ErrWeakPassword)
	}

	hash, err := u.hasher.Hash(input.Password)
	if err != nil {
		return inbound.RegisterUserOutput{}, fmt.Errorf("register user: %w", err)
	}

	newUser, err := user.New(u.idGen.NewID(), input.Email, hash, u.now())
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
