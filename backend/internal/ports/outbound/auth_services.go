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
