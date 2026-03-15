package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"micha/backend/internal/ports/outbound"
)

// ErrWeakJWTSecret is returned when the JWT secret is shorter than 32 bytes.
var ErrWeakJWTSecret = errors.New("JWT secret must be at least 32 characters")

// JWTSigner implements outbound.TokenSigner using HS256.
type JWTSigner struct{ secret []byte }

// NewJWTSigner constructs a JWTSigner with the given secret.
// Returns ErrWeakJWTSecret if the secret is shorter than 32 bytes.
func NewJWTSigner(secret string) (JWTSigner, error) {
	if len(secret) < 32 {
		return JWTSigner{}, ErrWeakJWTSecret
	}
	return JWTSigner{secret: []byte(secret)}, nil
}

// Sign creates a signed HS256 JWT containing the user's ID and email.
func (s JWTSigner) Sign(userID, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   now.Add(24 * time.Hour).Unix(),
		"iat":   now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return signed, nil
}

// JWTValidator validates HS256 tokens and extracts claims.
type JWTValidator struct{ secret []byte }

// NewJWTValidator constructs a JWTValidator with the given secret.
// Returns ErrWeakJWTSecret if the secret is shorter than 32 bytes.
func NewJWTValidator(secret string) (JWTValidator, error) {
	if len(secret) < 32 {
		return JWTValidator{}, ErrWeakJWTSecret
	}
	return JWTValidator{secret: []byte(secret)}, nil
}

// Validate parses and verifies a JWT, returning the userID and email from claims.
func (v JWTValidator) Validate(tokenString string) (userID, email string, err error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}

	userID, _ = claims["sub"].(string)
	email, _ = claims["email"].(string)
	return userID, email, nil
}

var _ outbound.TokenSigner = JWTSigner{}
var _ outbound.TokenValidator = JWTValidator{}
