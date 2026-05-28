package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/outbound"
)

// UserRepository fulfils outbound.UserRepository using PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository constructs a UserRepository backed by the given pool.
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return UserRepository{db: db}
}

// Save persists a new user record. Returns shared.ErrAlreadyExists on duplicate email.
func (r UserRepository) Save(ctx context.Context, u user.User) error {
	attrs := u.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, email, password_hash, name, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		attrs.ID, attrs.Email, attrs.PasswordHash, attrs.Name, attrs.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return shared.ErrAlreadyExists
		}
		return fmt.Errorf("user repository save: %w", err)
	}
	return nil
}

// FindByEmail retrieves a user by email. Returns shared.ErrNotFound when absent.
func (r UserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at
		 FROM users
		 WHERE email = $1`,
		email,
	)

	var attrs user.UserAttributes
	if err := row.Scan(&attrs.ID, &attrs.Email, &attrs.PasswordHash, &attrs.Name, &attrs.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, shared.ErrNotFound
		}
		return user.User{}, fmt.Errorf("user repository findByEmail: %w", err)
	}

	u, err := user.NewFromAttributes(attrs)
	if err != nil {
		return user.User{}, fmt.Errorf("user repository findByEmail: rehydrate: %w", err)
	}
	return u, nil
}

// FindByID retrieves a user by ID. Returns shared.ErrNotFound when absent.
func (r UserRepository) FindByID(ctx context.Context, id string) (user.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, name, created_at
		 FROM users
		 WHERE id = $1`,
		id,
	)

	var attrs user.UserAttributes
	if err := row.Scan(&attrs.ID, &attrs.Email, &attrs.PasswordHash, &attrs.Name, &attrs.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, shared.ErrNotFound
		}
		return user.User{}, fmt.Errorf("user repository findByID: %w", err)
	}

	u, err := user.NewFromAttributes(attrs)
	if err != nil {
		return user.User{}, fmt.Errorf("user repository findByID: rehydrate: %w", err)
	}
	return u, nil
}

var _ outbound.UserRepository = UserRepository{}
