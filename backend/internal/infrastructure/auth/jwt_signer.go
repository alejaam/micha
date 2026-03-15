package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"micha/backend/internal/ports/outbound"
)

// JWTSigner implements outbound.TokenSigner using HS256.
type JWTSigner struct{ secret []byte }

// NewJWTSigner constructs a JWTSigner with the given secret.
func NewJWTSigner(secret string) JWTSigner { return JWTSigner{secret: []byte(secret)} }

// Sign creates a signed HS256 JWT containing the user's ID and email.
func (s JWTSigner) Sign(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
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
func NewJWTValidator(secret string) JWTValidator { return JWTValidator{secret: []byte(secret)} }

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
