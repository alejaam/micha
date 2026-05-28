package user

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	// ErrInvalidEmail is returned when the email field is missing or malformed.
	ErrInvalidEmail = errors.New("invalid email")
	// ErrWeakPassword is returned when the password hash field is empty.
	ErrWeakPassword = errors.New("weak password")
	// ErrInvalidName is returned when the name field is invalid.
	ErrInvalidName = errors.New("invalid name")
)

// UserAttributes is the flat DTO used for construction and rehydration.
type UserAttributes struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	CreatedAt    time.Time
}

// User is the aggregate root for a user account.
type User struct {
	id           string
	email        string
	passwordHash string
	name         string
	createdAt    time.Time
}

// New constructs a User from individual fields.
func New(id, email, passwordHash, name string, createdAt time.Time) (User, error) {
	return NewFromAttributes(UserAttributes{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		CreatedAt:    createdAt,
	})
}

// NewFromAttributes constructs a User from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs UserAttributes) (User, error) {
	email := strings.TrimSpace(attrs.Email)
	if email == "" || !strings.Contains(email, "@") {
		return User{}, ErrInvalidEmail
	}

	if strings.TrimSpace(attrs.PasswordHash) == "" {
		return User{}, ErrWeakPassword
	}

	name := strings.TrimSpace(attrs.Name)
	nameLen := utf8.RuneCountInString(name)
	if name == "" || nameLen < 1 || nameLen > 100 {
		return User{}, ErrInvalidName
	}

	return User{
		id:           attrs.ID,
		email:        email,
		passwordHash: strings.TrimSpace(attrs.PasswordHash),
		name:         name,
		createdAt:    attrs.CreatedAt,
	}, nil
}

// Attributes returns the flat representation of the user.
func (u User) Attributes() UserAttributes {
	return UserAttributes{
		ID:           u.id,
		Email:        u.email,
		PasswordHash: u.passwordHash,
		Name:         u.name,
		CreatedAt:    u.createdAt,
	}
}

// ID returns the user's unique identifier.
func (u User) ID() string { return u.id }

// Email returns the user's email address.
func (u User) Email() string { return u.email }

// PasswordHash returns the user's hashed password.
func (u User) PasswordHash() string { return u.passwordHash }

// Name returns the user's name.
func (u User) Name() string { return u.name }

// CreatedAt returns the time the user account was created.
func (u User) CreatedAt() time.Time { return u.createdAt }
