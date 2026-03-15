// Package postgres provides PostgreSQL adapter implementations for outbound ports.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

// HouseholdRepository fulfils outbound.HouseholdRepository using PostgreSQL.
type HouseholdRepository struct {
	db *pgxpool.Pool
}

// NewHouseholdRepository constructs a HouseholdRepository backed by the given pool.
func NewHouseholdRepository(db *pgxpool.Pool) HouseholdRepository {
	return HouseholdRepository{db: db}
}

// Save persists a new household record.
func (r HouseholdRepository) Save(ctx context.Context, h household.Household) error {
	attrs := h.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO households (id, name, settlement_mode, currency, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
		string(attrs.ID), attrs.Name, string(attrs.SettlementMode), attrs.Currency, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("household repository save: %w", err)
	}

	return nil
}

// FindByID retrieves a household by ID.
func (r HouseholdRepository) FindByID(ctx context.Context, id string) (household.Household, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, settlement_mode, currency, created_at, updated_at
			FROM households
			WHERE id = $1`,
		id,
	)

	h, err := scanHousehold(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return household.Household{}, shared.ErrNotFound
	}
	if err != nil {
		return household.Household{}, fmt.Errorf("household repository findByID: %w", err)
	}

	return h, nil
}

// List returns households ordered by created_at DESC.
func (r HouseholdRepository) List(ctx context.Context, limit, offset int) ([]household.Household, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, settlement_mode, currency, created_at, updated_at
			FROM households
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("household repository list: %w", err)
	}
	defer rows.Close()

	var households []household.Household
	for rows.Next() {
		h, err := scanHousehold(rows)
		if err != nil {
			return nil, fmt.Errorf("household repository list: scan: %w", err)
		}
		households = append(households, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("household repository list: rows: %w", err)
	}

	return households, nil
}

// Update persists changes to an existing household.
func (r HouseholdRepository) Update(ctx context.Context, h household.Household) error {
	attrs := h.Attributes()
	tag, err := r.db.Exec(ctx,
		`UPDATE households
			SET name = $1,
				settlement_mode = $2,
				currency = $3,
				updated_at = $4
			WHERE id = $5`,
		attrs.Name, string(attrs.SettlementMode), attrs.Currency, attrs.UpdatedAt, string(attrs.ID),
	)
	if err != nil {
		return fmt.Errorf("household repository update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// ensure interface compliance at compile time.
var _ outbound.HouseholdRepository = HouseholdRepository{}

// ListByUserID returns households that the given user belongs to (via members table).
func (r HouseholdRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]household.Household, error) {
	rows, err := r.db.Query(ctx,
		`SELECT h.id, h.name, h.settlement_mode, h.currency, h.created_at, h.updated_at
			FROM households h
			INNER JOIN members m ON m.household_id = h.id
			WHERE m.user_id = $1
			ORDER BY h.created_at DESC
			LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("household repository listByUserID: %w", err)
	}
	defer rows.Close()

	var households []household.Household
	for rows.Next() {
		h, err := scanHousehold(rows)
		if err != nil {
			return nil, fmt.Errorf("household repository listByUserID: scan: %w", err)
		}
		households = append(households, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("household repository listByUserID: rows: %w", err)
	}

	return households, nil
}

func scanHousehold(r row) (household.Household, error) {
	var (
		id             string
		name           string
		settlementMode string
		currency       string
		createdAt      time.Time
		updatedAt      time.Time
	)

	if err := r.Scan(&id, &name, &settlementMode, &currency, &createdAt, &updatedAt); err != nil {
		return household.Household{}, err
	}

	return household.NewFromAttributes(household.Attributes{
		ID:             household.ID(id),
		Name:           name,
		SettlementMode: household.SettlementMode(settlementMode),
		Currency:       currency,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	})
}
