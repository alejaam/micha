package outbound

import (
	"context"

	"micha/backend/internal/domain/pushsubscription"
)

// PushSubscriptionRepository defines the contract for persisting push subscriptions.
type PushSubscriptionRepository interface {
	// Save persists a push subscription, replacing any existing one with the same user_id + endpoint.
	Save(ctx context.Context, sub pushsubscription.PushSubscription) error

	// ListByUserID returns all push subscriptions for a given user.
	ListByUserID(ctx context.Context, userID string) ([]pushsubscription.PushSubscription, error)

	// Delete removes a push subscription by its endpoint (for when a subscription expires).
	DeleteByEndpoint(ctx context.Context, userID string, endpoint string) error
}
