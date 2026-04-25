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

// loadSplitConfig retrieves the split configuration rows for a household.
// Returns an empty SplitConfig (equal-split default) if no rows exist.
func loadSplitConfig(ctx context.Context, db *pgxpool.Pool, householdID string) (household.SplitConfig, error) {
	rows, err := db.Query(ctx,
		`SELECT member_id, percentage FROM household_split_config WHERE household_id = $1 ORDER BY member_id`,
		householdID,
	)
	if err != nil {
		return household.SplitConfig{}, fmt.Errorf("load split config: %w", err)
	}
	defer rows.Close()

	var splits []household.MemberSplit
	for rows.Next() {
		var memberID string
		var pct float64
		if err := rows.Scan(&memberID, &pct); err != nil {
			return household.SplitConfig{}, fmt.Errorf("load split config: scan: %w", err)
		}
		splits = append(splits, household.MemberSplit{MemberID: memberID, Percentage: pct})
	}
	if err := rows.Err(); err != nil {
		return household.SplitConfig{}, fmt.Errorf("load split config: rows: %w", err)
	}

	if len(splits) == 0 {
		return household.SplitConfig{}, nil
	}
	return household.NewSplitConfig(splits)
}

// saveSplitConfig replaces all split config rows for a household inside the given transaction.
func saveSplitConfig(ctx context.Context, db *pgxpool.Pool, householdID string, sc household.SplitConfig) error {
	if _, err := db.Exec(ctx,
		`DELETE FROM household_split_config WHERE household_id = $1`,
		householdID,
	); err != nil {
		return fmt.Errorf("save split config delete: %w", err)
	}

	for _, s := range sc.Splits() {
		if _, err := db.Exec(ctx,
			`INSERT INTO household_split_config (household_id, member_id, percentage) VALUES ($1, $2, $3)`,
			householdID, s.MemberID, s.Percentage,
		); err != nil {
			return fmt.Errorf("save split config insert: %w", err)
		}
	}
	return nil
}

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
		`INSERT INTO households (id, name, owner_id, settlement_mode, currency, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		string(attrs.ID), attrs.Name, nullIfEmpty(attrs.OwnerID), string(attrs.SettlementMode), attrs.Currency, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("household repository save: %w", err)
	}

	return nil
}

// FindByID retrieves a household by ID.
func (r HouseholdRepository) FindByID(ctx context.Context, id string) (household.Household, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, name, owner_id, settlement_mode, currency, created_at, updated_at
			FROM households
			WHERE id = $1`,
		id,
	)

	h, err := scanHousehold(row)
	if err != nil {
		// FALLBACK: If owner_id column is missing, try legacy query
		legacyRow := r.db.QueryRow(ctx,
			`SELECT id, name, settlement_mode, currency, created_at, updated_at
				FROM households
				WHERE id = $1`,
			id,
		)
		var (
			lid, name, sm, curr string
			ca, ua              time.Time
		)
		if lerr := legacyRow.Scan(&lid, &name, &sm, &curr, &ca, &ua); lerr == nil {
			return household.NewFromAttributes(household.Attributes{
				ID:             household.ID(lid),
				Name:           name,
				OwnerID:        "",
				SettlementMode: household.SettlementMode(sm),
				Currency:       curr,
				CreatedAt:      ca,
				UpdatedAt:      ua,
			})
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return household.Household{}, shared.ErrNotFound
		}
		return household.Household{}, fmt.Errorf("household repository findByID: %w", err)
	}

	sc, err := loadSplitConfig(ctx, r.db, id)
	if err != nil {
		return household.Household{}, fmt.Errorf("household repository findByID: %w", err)
	}
	attrs := h.Attributes()
	attrs.SplitConfig = sc
	return household.NewFromAttributes(attrs)
}

// List returns households ordered by created_at DESC.
func (r HouseholdRepository) List(ctx context.Context, limit, offset int) ([]household.Household, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, owner_id, settlement_mode, currency, created_at, updated_at
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

	if !attrs.SplitConfig.IsEmpty() {
		if err := saveSplitConfig(ctx, r.db, string(attrs.ID), attrs.SplitConfig); err != nil {
			return fmt.Errorf("household repository update: %w", err)
		}
	}

	return nil
}

// ensure interface compliance at compile time.
var _ outbound.HouseholdRepository = HouseholdRepository{}

// ListByUserID returns households that the given user belongs to (via members table).
func (r HouseholdRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]household.Household, error) {
	rows, err := r.db.Query(ctx,
		`SELECT h.id, h.name, h.owner_id, h.settlement_mode, h.currency, h.created_at, h.updated_at
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
		ownerID        *string
		settlementMode string
		currency       string
		createdAt      time.Time
		updatedAt      time.Time
	)

	if err := r.Scan(&id, &name, &ownerID, &settlementMode, &currency, &createdAt, &updatedAt); err != nil {
		return household.Household{}, err
	}

	return household.NewFromAttributes(household.Attributes{
		ID:             household.ID(id),
		Name:           name,
		OwnerID:        valueOrEmptyString(ownerID),
		SettlementMode: household.SettlementMode(settlementMode),
		Currency:       currency,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	})
}
