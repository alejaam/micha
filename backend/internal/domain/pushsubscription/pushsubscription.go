// Package pushsubscription holds the domain entity for Web Push subscriptions.
package pushsubscription

import (
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// ID is the unique identifier type for a push subscription.
type ID string

// Attributes is the flat DTO for construction and rehydration.
type Attributes struct {
	ID        ID
	UserID    string
	Endpoint  string
	P256dhKey string
	AuthKey   string
	CreatedAt time.Time
}

// PushSubscription represents a browser push subscription registered by a user.
type PushSubscription struct {
	id        ID
	userID    string
	endpoint  string
	p256dhKey string
	authKey   string
	createdAt time.Time
}

// New constructs a PushSubscription from individual fields.
func New(id ID, userID string, endpoint string, p256dhKey string, authKey string, createdAt time.Time) (PushSubscription, error) {
	return NewFromAttributes(Attributes{
		ID:        id,
		UserID:    userID,
		Endpoint:  endpoint,
		P256dhKey: p256dhKey,
		AuthKey:   authKey,
		CreatedAt: createdAt,
	})
}

// NewFromAttributes constructs a PushSubscription from a flat attribute bag.
func NewFromAttributes(attrs Attributes) (PushSubscription, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return PushSubscription{}, shared.ErrInvalidID
	}
	if strings.TrimSpace(attrs.UserID) == "" {
		return PushSubscription{}, shared.ErrInvalidID
	}
	endpoint := strings.TrimSpace(attrs.Endpoint)
	if endpoint == "" {
		return PushSubscription{}, shared.ErrInvalidName
	}
	p256dhKey := strings.TrimSpace(attrs.P256dhKey)
	if p256dhKey == "" {
		return PushSubscription{}, shared.ErrInvalidName
	}
	authKey := strings.TrimSpace(attrs.AuthKey)
	if authKey == "" {
		return PushSubscription{}, shared.ErrInvalidName
	}

	createdAt := attrs.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return PushSubscription{
		id:        attrs.ID,
		userID:    attrs.UserID,
		endpoint:  endpoint,
		p256dhKey: p256dhKey,
		authKey:   authKey,
		createdAt: createdAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (s PushSubscription) Attributes() Attributes {
	return Attributes{
		ID:        s.id,
		UserID:    s.userID,
		Endpoint:  s.endpoint,
		P256dhKey: s.p256dhKey,
		AuthKey:   s.authKey,
		CreatedAt: s.createdAt,
	}
}

func (s PushSubscription) ID() ID             { return s.id }
func (s PushSubscription) UserID() string      { return s.userID }
func (s PushSubscription) Endpoint() string    { return s.endpoint }
func (s PushSubscription) P256dhKey() string   { return s.p256dhKey }
func (s PushSubscription) AuthKey() string     { return s.authKey }
func (s PushSubscription) CreatedAt() time.Time { return s.createdAt }
