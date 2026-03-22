package outbound

import (
	"context"

	"micha/backend/internal/domain/user"
)

// UserRepository defines the persistence contract for user data.
type UserRepository interface {
	Save(ctx context.Context, u user.User) error
	FindByEmail(ctx context.Context, email string) (user.User, error)
}
