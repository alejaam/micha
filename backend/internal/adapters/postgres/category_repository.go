package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/outbound"
)

// CategoryRepository fulfils outbound.CategoryRepository using PostgreSQL.
type CategoryRepository struct {
	db *pgxpool.Pool
}

// NewCategoryRepository constructs a CategoryRepository backed by the given pool.
func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return CategoryRepository{db: db}
}

// Save persists a new category record.
func (r CategoryRepository) Save(ctx context.Context, c category.Category) error {
	attrs := c.Attributes()
	_, err := r.db.Exec(ctx,
		`INSERT INTO categories (id, household_id, name, slug, created_at)
			VALUES ($1, $2, $3, $4, $5)`,
		string(attrs.ID), attrs.HouseholdID, attrs.Name, attrs.Slug, attrs.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("category repository save: %w", err)
	}
	return nil
}

// FindBySlug retrieves a category by household and slug.
func (r CategoryRepository) FindBySlug(ctx context.Context, householdID, slug string) (category.Category, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, household_id, name, slug, created_at
			FROM categories
			WHERE household_id = $1 AND slug = $2
			LIMIT 1`,
		householdID, slug,
	)

	c, err := scanCategory(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return category.Category{}, shared.ErrNotFound
	}
	if err != nil {
		return category.Category{}, fmt.Errorf("category repository findBySlug: %w", err)
	}
	return c, nil
}

// ListByHousehold returns all categories for a household ordered by name ASC.
func (r CategoryRepository) ListByHousehold(ctx context.Context, householdID string) ([]category.Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, household_id, name, slug, created_at
			FROM categories
			WHERE household_id = $1
			ORDER BY name ASC`,
		householdID,
	)
	if err != nil {
		return nil, fmt.Errorf("category repository listByHousehold: %w", err)
	}
	defer rows.Close()

	var cats []category.Category
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, fmt.Errorf("category repository listByHousehold: scan: %w", err)
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category repository listByHousehold: rows: %w", err)
	}
	return cats, nil
}

// Delete removes a category by ID.
func (r CategoryRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM categories WHERE id = $1 AND is_default = false`,
		id,
	)
	if err != nil {
		return fmt.Errorf("category repository delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// ensure interface compliance at compile time.
var _ outbound.CategoryRepository = CategoryRepository{}

func scanCategory(r row) (category.Category, error) {
	var (
		id          string
		householdID string
		name        string
		slug        string
		createdAt   time.Time
	)

	if err := r.Scan(&id, &householdID, &name, &slug, &createdAt); err != nil {
		return category.Category{}, err
	}

	return category.NewFromAttributes(category.Attributes{
		ID:          category.ID(id),
		HouseholdID: householdID,
		Name:        name,
		Slug:        slug,
		CreatedAt:   createdAt,
	})
}
