package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/pushsubscription"
)

// PushSubscriptionRepository fulfils outbound.PushSubscriptionRepository using PostgreSQL.
type PushSubscriptionRepository struct {
	db *pgxpool.Pool
}

// NewPushSubscriptionRepository constructs a PushSubscriptionRepository backed by the given pool.
func NewPushSubscriptionRepository(db *pgxpool.Pool) PushSubscriptionRepository {
	return PushSubscriptionRepository{db: db}
}

// Save persists a push subscription (upserts on user_id + endpoint conflict).
func (r PushSubscriptionRepository) Save(ctx context.Context, sub pushsubscription.PushSubscription) error {
	attrs := sub.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO push_subscriptions (id, user_id, endpoint, p256dh_key, auth_key, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (user_id, endpoint) DO UPDATE SET
		   p256dh_key = EXCLUDED.p256dh_key,
		   auth_key = EXCLUDED.auth_key`,
		string(attrs.ID), attrs.UserID, attrs.Endpoint, attrs.P256dhKey, attrs.AuthKey, attrs.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("push subscription repository save: %w", err)
	}
	return nil
}

// ListByUserID returns all push subscriptions for a given user.
func (r PushSubscriptionRepository) ListByUserID(ctx context.Context, userID string) ([]pushsubscription.PushSubscription, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, endpoint, p256dh_key, auth_key, created_at
		 FROM push_subscriptions
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("push subscription repository listByUserID: %w", err)
	}
	defer rows.Close()

	var subs []pushsubscription.PushSubscription
	for rows.Next() {
		var id, userID, endpoint, p256dh, auth string
		var createdAt time.Time

		if err := rows.Scan(&id, &userID, &endpoint, &p256dh, &auth, &createdAt); err != nil {
			return nil, fmt.Errorf("push subscription repository scan: %w", err)
		}

		sub, err := pushsubscription.New(
			pushsubscription.ID(id), userID, endpoint, p256dh, auth, createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("push subscription repository rehydrate: %w", err)
		}
		subs = append(subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("push subscription repository rows: %w", err)
	}

	return subs, nil
}

// DeleteByEndpoint removes a push subscription by endpoint for a given user.
func (r PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, userID string, endpoint string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM push_subscriptions WHERE user_id = $1 AND endpoint = $2`,
		userID, endpoint,
	)
	if err != nil {
		return fmt.Errorf("push subscription repository delete: %w", err)
	}
	return nil
}
