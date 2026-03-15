// Package category holds the Category domain entity.
package category

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidName        = errors.New("invalid category name")
	ErrInvalidSlug        = errors.New("invalid category slug: must be lowercase alphanumeric with hyphens, max 64 chars")
	ErrInvalidHouseholdID = errors.New("invalid category household id")
)

// slugPattern validates that slugs are lowercase alphanumeric + hyphens.
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
	HouseholdID string
	Name        string
	Slug        string
	IsDefault   bool
	CreatedAt   time.Time
}

// Category is the domain entity representing an expense category.
// Default categories (seeded at household creation) have IsDefault=true
// and cannot be deleted by users.
type Category struct {
	id          ID
	householdID string
	name        string
	slug        string
	isDefault   bool
	createdAt   time.Time
}

// New constructs a custom (non-default) Category.
func New(id ID, householdID, name, slug string, createdAt time.Time) (Category, error) {
	return NewFromAttributes(Attributes{
		ID:          id,
		HouseholdID: householdID,
		Name:        name,
		Slug:        slug,
		IsDefault:   false,
		CreatedAt:   createdAt,
	})
}

// NewFromAttributes constructs a Category from a flat attribute bag.
func NewFromAttributes(attrs Attributes) (Category, error) {
	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return Category{}, ErrInvalidHouseholdID
	}

	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Category{}, ErrInvalidName
	}

	slug := strings.TrimSpace(attrs.Slug)
	if slug == "" || len(slug) > 64 || !slugPattern.MatchString(slug) {
		return Category{}, ErrInvalidSlug
	}

	return Category{
		id:          attrs.ID,
		householdID: attrs.HouseholdID,
		name:        name,
		slug:        slug,
		isDefault:   attrs.IsDefault,
		createdAt:   attrs.CreatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (c Category) Attributes() Attributes {
	return Attributes{
		ID:          c.id,
		HouseholdID: c.householdID,
		Name:        c.name,
		Slug:        c.slug,
		IsDefault:   c.isDefault,
		CreatedAt:   c.createdAt,
	}
}

func (c Category) ID() ID               { return c.id }
func (c Category) HouseholdID() string  { return c.householdID }
func (c Category) Name() string         { return c.name }
func (c Category) Slug() string         { return c.slug }
func (c Category) IsDefault() bool      { return c.isDefault }
func (c Category) CreatedAt() time.Time { return c.createdAt }
