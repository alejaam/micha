package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	authapp "micha/backend/internal/application/auth"
)

// BcryptHasher implements authapp.PasswordHasher using bcrypt at cost 12.
type BcryptHasher struct{}

// NewBcryptHasher constructs a BcryptHasher.
func NewBcryptHasher() BcryptHasher { return BcryptHasher{} }

// Hash generates a bcrypt hash of the given password.
func (BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// Verify checks whether password matches the stored hash.
func (BcryptHasher) Verify(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

var _ authapp.PasswordHasher = BcryptHasher{}
