package inbound

import "context"

// RegisterUserInput contains the data required to create a new user account.
type RegisterUserInput struct {
	Email    string
	Password string
}

// RegisterUserOutput contains the new user's ID.
type RegisterUserOutput struct {
	UserID string
}

// LoginInput contains the credentials for authentication.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput contains the signed JWT access token.
type LoginOutput struct {
	Token string
}

// RegisterUserUseCase creates a new user account.
type RegisterUserUseCase interface {
	Execute(ctx context.Context, input RegisterUserInput) (RegisterUserOutput, error)
}

// LoginUseCase authenticates a user and returns a JWT token.
type LoginUseCase interface {
	Execute(ctx context.Context, input LoginInput) (LoginOutput, error)
}
