// Package category holds the Category domain entity.
package category

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

var (
	ErrInvalidSlug = errors.New("invalid category slug")
)

// slugPattern validates that slugs are lowercase alphanumeric + hyphens (no leading/trailing hyphens).
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// DefaultSlugs are the built-in categories pre-seeded for every household.
var DefaultSlugs = []string{
	"rent", "auto", "streaming", "food", "personal", "savings", "other",
}

// ID is the unique identifier for a Category.
type ID string

// Attributes is the flat DTO for construction and rehydration.
type Attributes struct {
	ID          ID
	Name        string
	Slug        string // lowercase, alphanumeric + hyphens only
	HouseholdID string // empty if predefined/default
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Category is the domain entity representing an expense category.
// Default categories have empty HouseholdID and their slug is in DefaultSlugs.
// Custom categories belong to a specific household and have a custom slug.
type Category struct {
	id          ID
	name        string
	slug        string
	householdID string
	createdAt   time.Time
	updatedAt   time.Time
}

// New constructs a Category from individual fields.
func New(id string, householdID string, name string, slug string, createdAt time.Time) (Category, error) {
	return NewFromAttributes(Attributes{
		ID:          ID(id),
		Name:        name,
		Slug:        slug,
		HouseholdID: householdID,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	})
}

// NewFromAttributes constructs a Category from a flat attribute bag.
func NewFromAttributes(attrs Attributes) (Category, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Category{}, shared.ErrInvalidID
	}

	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Category{}, shared.ErrInvalidName
	}

	slug := strings.TrimSpace(attrs.Slug)
	if slug == "" || slug != strings.ToLower(slug) || !slugPattern.MatchString(slug) {
		return Category{}, ErrInvalidSlug
	}

	householdID := strings.TrimSpace(attrs.HouseholdID)

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Category{
		id:          attrs.ID,
		name:        name,
		slug:        slug,
		householdID: householdID,
		createdAt:   attrs.CreatedAt,
		updatedAt:   updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (c Category) Attributes() Attributes {
	return Attributes{
		ID:          c.id,
		Name:        c.name,
		Slug:        c.slug,
		HouseholdID: c.householdID,
		CreatedAt:   c.createdAt,
		UpdatedAt:   c.updatedAt,
	}
}

func (c Category) ID() ID               { return c.id }
func (c Category) Name() string         { return c.name }
func (c Category) Slug() string         { return c.slug }
func (c Category) HouseholdID() string  { return c.householdID }
func (c Category) CreatedAt() time.Time { return c.createdAt }
func (c Category) UpdatedAt() time.Time { return c.updatedAt }

// IsDefault returns true if this is a default category (no household association).
func (c Category) IsDefault() bool {
	for _, d := range DefaultSlugs {
		if c.slug == d {
			return true
		}
	}
	return false
}
