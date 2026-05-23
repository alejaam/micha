// Package subscriptioncatalog provides domain entities for the subscription catalog.
package subscriptioncatalog

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidName            = errors.New("service name is required")
	ErrInvalidSlug            = errors.New("service slug is required")
	ErrInvalidPrice           = errors.New("standalone price must be greater than zero")
	ErrInvalidCustomPrice     = errors.New("custom price must be greater than zero when provided")
	ErrCatalogServiceNotFound = errors.New("catalog service not found")
	ErrLinkAlreadyExists      = errors.New("service already linked to this recurring expense")
	ErrLinkNotFound           = errors.New("catalog link not found")
	ErrInvalidID              = errors.New("invalid id")
)

// ID is the unique identifier for a subscription service.
type ID string

// SubscriptionService represents a catalog entry for a streaming/bundle service.
type SubscriptionService struct {
	id                   ID
	name                 string
	slug                 string
	region               string
	currency             string
	standalonePriceCents int64
	iconURL              string
	isBundle             bool
	createdAt            time.Time
	updatedAt            time.Time
}

// SubscriptionServiceAttributes is the flat DTO for construction and rehydration.
type SubscriptionServiceAttributes struct {
	ID                   ID
	Name                 string
	Slug                 string
	Region               string
	Currency             string
	StandalonePriceCents int64
	IconURL              string
	IsBundle             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewSubscriptionService constructs a SubscriptionService from individual fields.
func NewSubscriptionService(
	id ID,
	name, slug, region, currency string,
	standalonePriceCents int64,
	iconURL string,
	isBundle bool,
	createdAt time.Time,
) (SubscriptionService, error) {
	return NewSubscriptionServiceFromAttributes(SubscriptionServiceAttributes{
		ID:                   id,
		Name:                 name,
		Slug:                 slug,
		Region:               region,
		Currency:             currency,
		StandalonePriceCents: standalonePriceCents,
		IconURL:              iconURL,
		IsBundle:             isBundle,
		CreatedAt:            createdAt,
		UpdatedAt:            createdAt,
	})
}

// NewSubscriptionServiceFromAttributes constructs a SubscriptionService from a flat DTO.
func NewSubscriptionServiceFromAttributes(attrs SubscriptionServiceAttributes) (SubscriptionService, error) {
	if strings.TrimSpace(attrs.Name) == "" {
		return SubscriptionService{}, ErrInvalidName
	}
	if strings.TrimSpace(attrs.Slug) == "" {
		return SubscriptionService{}, ErrInvalidSlug
	}
	if attrs.StandalonePriceCents <= 0 {
		return SubscriptionService{}, ErrInvalidPrice
	}

	region := attrs.Region
	if region == "" {
		region = "MX"
	}
	currency := strings.ToUpper(attrs.Currency)
	if currency == "" {
		currency = "MXN"
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return SubscriptionService{
		id:                   attrs.ID,
		name:                 strings.TrimSpace(attrs.Name),
		slug:                 strings.TrimSpace(attrs.Slug),
		region:               region,
		currency:             currency,
		standalonePriceCents: attrs.StandalonePriceCents,
		iconURL:              attrs.IconURL,
		isBundle:             attrs.IsBundle,
		createdAt:            attrs.CreatedAt,
		updatedAt:            updatedAt,
	}, nil
}

// Attributes returns a flat DTO copy of the entity.
func (s SubscriptionService) Attributes() SubscriptionServiceAttributes {
	return SubscriptionServiceAttributes{
		ID:                   s.id,
		Name:                 s.name,
		Slug:                 s.slug,
		Region:               s.region,
		Currency:             s.currency,
		StandalonePriceCents: s.standalonePriceCents,
		IconURL:              s.iconURL,
		IsBundle:             s.isBundle,
		CreatedAt:            s.createdAt,
		UpdatedAt:            s.updatedAt,
	}
}

// ID returns the service ID.
func (s SubscriptionService) ID() ID                      { return s.id }
func (s SubscriptionService) Name() string                { return s.name }
func (s SubscriptionService) Slug() string                { return s.slug }
func (s SubscriptionService) Region() string              { return s.region }
func (s SubscriptionService) Currency() string            { return s.currency }
func (s SubscriptionService) StandalonePriceCents() int64 { return s.standalonePriceCents }
func (s SubscriptionService) IconURL() string             { return s.iconURL }
func (s SubscriptionService) IsBundle() bool              { return s.isBundle }
func (s SubscriptionService) CreatedAt() time.Time        { return s.createdAt }
func (s SubscriptionService) UpdatedAt() time.Time        { return s.updatedAt }
