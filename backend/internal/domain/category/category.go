// Package category holds the Category domain entity.
package category

import (
	"regexp"
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// CategoryType defines whether a category is predefined or custom.
type CategoryType string

const (
	CategoryTypePredefined CategoryType = "predefined"
	CategoryTypeCustom     CategoryType = "custom"
)

// slugPattern validates that slugs are lowercase alphanumeric + hyphens.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// DefaultSlugs are the built-in categories pre-seeded for every household.
var DefaultSlugs = []string{
	"rent", "auto", "streaming", "food", "personal", "savings", "other",
}

// ID is the unique identifier for a Category.
type ID string

// CategoryAttributes is the flat DTO for construction and rehydration.
type CategoryAttributes struct {
	ID           ID
	Name         string
	CategoryType CategoryType
	HouseholdID  string // empty/null if predefined
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Category is the domain entity representing an expense category.
// Predefined categories have HouseholdID empty and CategoryType=predefined.
// Custom categories belong to a specific household.
type Category struct {
	id           ID
	name         string
	categoryType CategoryType
	householdID  string
	createdAt    time.Time
	updatedAt    time.Time
}

// New constructs a Category from individual fields.
func New(id ID, name string, categoryType CategoryType, householdID string, createdAt time.Time) (Category, error) {
	return NewFromAttributes(CategoryAttributes{
		ID:           id,
		Name:         name,
		CategoryType: categoryType,
		HouseholdID:  householdID,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	})
}

// NewFromAttributes constructs a Category from a flat attribute bag.
func NewFromAttributes(attrs CategoryAttributes) (Category, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Category{}, shared.ErrInvalidID
	}

	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Category{}, shared.ErrInvalidName
	}

	categoryType := attrs.CategoryType
	if categoryType == "" {
		categoryType = CategoryTypeCustom
	}
	if categoryType != CategoryTypePredefined && categoryType != CategoryTypeCustom {
		return Category{}, shared.ErrInvalidStatus
	}

	// Predefined categories should have empty householdID
	if categoryType == CategoryTypePredefined && strings.TrimSpace(attrs.HouseholdID) != "" {
		return Category{}, shared.ErrInvalidStatus
	}

	// Custom categories must have a householdID
	if categoryType == CategoryTypeCustom && strings.TrimSpace(attrs.HouseholdID) == "" {
		return Category{}, shared.ErrInvalidID
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Category{
		id:           attrs.ID,
		name:         name,
		categoryType: categoryType,
		householdID:  attrs.HouseholdID,
		createdAt:    attrs.CreatedAt,
		updatedAt:    updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (c Category) Attributes() CategoryAttributes {
	return CategoryAttributes{
		ID:           c.id,
		Name:         c.name,
		CategoryType: c.categoryType,
		HouseholdID:  c.householdID,
		CreatedAt:    c.createdAt,
		UpdatedAt:    c.updatedAt,
	}
}

func (c Category) ID() ID                     { return c.id }
func (c Category) Name() string               { return c.name }
func (c Category) CategoryType() CategoryType { return c.categoryType }
func (c Category) HouseholdID() string        { return c.householdID }
func (c Category) CreatedAt() time.Time       { return c.createdAt }
func (c Category) UpdatedAt() time.Time       { return c.updatedAt }
