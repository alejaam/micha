package outbound

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// TokenSigner signs a user ID and email into a JWT token.
type TokenSigner interface {
	Sign(userID, email string) (string, error)
}

// TokenValidator validates a Bearer JWT and returns the embedded claims.
type TokenValidator interface {
	Validate(tokenString string) (userID, email string, err error)
}
