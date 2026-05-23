package subscriptioncatalog

import (
	"time"
)

// CatalogLink represents the link between a recurring expense and a catalog service.
type CatalogLink struct {
	recurringExpenseID string
	catalogServiceID   string
	customPriceCents   *int64 // nil means use catalog price
	createdAt          time.Time
}

// CatalogLinkAttributes is the flat DTO for construction and rehydration.
type CatalogLinkAttributes struct {
	RecurringExpenseID string
	CatalogServiceID   string
	CustomPriceCents   *int64
	CreatedAt          time.Time
}

// NewCatalogLink constructs a CatalogLink from attributes.
func NewCatalogLink(attrs CatalogLinkAttributes) (CatalogLink, error) {
	if attrs.RecurringExpenseID == "" {
		return CatalogLink{}, ErrInvalidID
	}
	if attrs.CatalogServiceID == "" {
		return CatalogLink{}, ErrInvalidID
	}
	if attrs.CustomPriceCents != nil && *attrs.CustomPriceCents <= 0 {
		return CatalogLink{}, ErrInvalidCustomPrice
	}

	createdAt := attrs.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return CatalogLink{
		recurringExpenseID: attrs.RecurringExpenseID,
		catalogServiceID:   attrs.CatalogServiceID,
		customPriceCents:   attrs.CustomPriceCents,
		createdAt:          createdAt,
	}, nil
}

// Attributes returns a flat DTO copy.
func (l CatalogLink) Attributes() CatalogLinkAttributes {
	return CatalogLinkAttributes{
		RecurringExpenseID: l.recurringExpenseID,
		CatalogServiceID:   l.catalogServiceID,
		CustomPriceCents:   l.customPriceCents,
		CreatedAt:          l.createdAt,
	}
}

func (l CatalogLink) RecurringExpenseID() string { return l.recurringExpenseID }
func (l CatalogLink) CatalogServiceID() string   { return l.catalogServiceID }
func (l CatalogLink) CustomPriceCents() *int64   { return l.customPriceCents }
func (l CatalogLink) CreatedAt() time.Time       { return l.createdAt }
