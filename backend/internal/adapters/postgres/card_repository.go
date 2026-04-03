package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/card"
	"micha/backend/internal/domain/shared"
)

// CardRepository fulfils outbound.CardRepository using PostgreSQL.
type CardRepository struct {
	db *pgxpool.Pool
}

// NewCardRepository constructs a CardRepository backed by the given pool.
func NewCardRepository(db *pgxpool.Pool) CardRepository {
	return CardRepository{db: db}
}

// Save persists a new card record.
func (r CardRepository) Save(ctx context.Context, c card.Card) error {
	attrs := c.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO cards (id, household_id, bank_name, card_name, cutoff_day, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		string(attrs.ID), attrs.HouseholdID, attrs.BankName, attrs.CardName, attrs.CutoffDay, attrs.CreatedAt, attrs.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return shared.ErrAlreadyExists
		}
		return fmt.Errorf("card repository save: %w", err)
	}

	return nil
}

// FindByID retrieves a card by ID.
func (r CardRepository) FindByID(ctx context.Context, id string) (card.Card, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, bank_name, card_name, cutoff_day, created_at, updated_at, deleted_at
			FROM cards
			WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)

	c, err := scanCard(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return card.Card{}, shared.ErrNotFound
	}
	if err != nil {
		return card.Card{}, fmt.Errorf("card repository findByID: %w", err)
	}

	return c, nil
}

// ListByHousehold returns all active cards for a household ordered by created_at DESC.
func (r CardRepository) ListByHousehold(ctx context.Context, householdID string) ([]card.Card, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, bank_name, card_name, cutoff_day, created_at, updated_at, deleted_at
			FROM cards
			WHERE household_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC`,
		householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("card repository listByHousehold: %w", err)
	}
	defer rows.Close()

	var cards []card.Card
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			return nil, fmt.Errorf("card repository listByHousehold: scan: %w", err)
		}
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("card repository listByHousehold: rows: %w", err)
	}

	return cards, nil
}

// Delete soft-deletes a card by setting deleted_at.
func (r CardRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	tag, err := r.db.Exec(ctx,
		`UPDATE cards
			SET deleted_at = $1, updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL`,
		now, id,
	)
	if err != nil {
		return fmt.Errorf("card repository delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}

	return nil
}

// scanner is an interface that both pgx.Row and pgx.Rows implement.
type scanner interface {
	Scan(dest ...any) error
}

// scanCard reconstructs a card.Card from a row/rows scanner.
func scanCard(r scanner) (card.Card, error) {
	var (
		id          string
		householdID string
		bankName    string
		cardName    string
		cutoffDay   int
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
	)

	if err := r.Scan(&id, &householdID, &bankName, &cardName, &cutoffDay, &createdAt, &updatedAt, &deletedAt); err != nil {
		return card.Card{}, err
	}

	return card.NewFromAttributes(card.Attributes{
		ID:          card.ID(id),
		HouseholdID: householdID,
		BankName:    bankName,
		CardName:    cardName,
		CutoffDay:   cutoffDay,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		DeletedAt:   deletedAt,
	})
}
