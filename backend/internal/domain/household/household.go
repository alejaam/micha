package household

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidName           = errors.New("invalid household name")
	ErrInvalidSettlementMode = errors.New("invalid settlement mode")
	ErrInvalidCurrency       = errors.New("invalid currency")
)

// ID is the unique identifier type for a household.
type ID string

// SettlementMode defines how shared expenses are split.
type SettlementMode string

const (
	SettlementModeEqual        SettlementMode = "equal"
	SettlementModeProportional SettlementMode = "proportional"
)

// Attributes is the flat DTO used for construction and rehydration.
type Attributes struct {
	ID             ID
	Name           string
	SettlementMode SettlementMode
	Currency       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Household is the aggregate root for a household.
type Household struct {
	id             ID
	name           string
	settlementMode SettlementMode
	currency       string
	createdAt      time.Time
	updatedAt      time.Time
}

// New constructs a Household from individual fields.
func New(id ID, name string, settlementMode SettlementMode, currency string, createdAt time.Time) (Household, error) {
	return NewFromAttributes(Attributes{
		ID:             id,
		Name:           name,
		SettlementMode: settlementMode,
		Currency:       currency,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})
}

// NewFromAttributes constructs a Household from a flat attribute bag.
func NewFromAttributes(attrs Attributes) (Household, error) {
	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Household{}, ErrInvalidName
	}

	if attrs.SettlementMode != SettlementModeEqual && attrs.SettlementMode != SettlementModeProportional {
		return Household{}, ErrInvalidSettlementMode
	}

	currency := strings.ToUpper(strings.TrimSpace(attrs.Currency))
	if len(currency) != 3 {
		return Household{}, ErrInvalidCurrency
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Household{
		id:             attrs.ID,
		name:           name,
		settlementMode: attrs.SettlementMode,
		currency:       currency,
		createdAt:      attrs.CreatedAt,
		updatedAt:      updatedAt,
	}, nil
}

// UpdateConfig updates mutable household fields.
func (h *Household) UpdateConfig(name string, settlementMode SettlementMode, currency string) error {
	candidate, err := NewFromAttributes(Attributes{
		ID:             h.id,
		Name:           name,
		SettlementMode: settlementMode,
		Currency:       currency,
		CreatedAt:      h.createdAt,
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		return err
	}

	h.name = candidate.name
	h.settlementMode = candidate.settlementMode
	h.currency = candidate.currency
	h.updatedAt = candidate.updatedAt
	return nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (h Household) Attributes() Attributes {
	return Attributes{
		ID:             h.id,
		Name:           h.name,
		SettlementMode: h.settlementMode,
		Currency:       h.currency,
		CreatedAt:      h.createdAt,
		UpdatedAt:      h.updatedAt,
	}
}

func (h Household) ID() ID                     { return h.id }
func (h Household) Name() string               { return h.name }
func (h Household) SettlementMode() SettlementMode { return h.settlementMode }
func (h Household) Currency() string           { return h.currency }
func (h Household) CreatedAt() time.Time       { return h.createdAt }
func (h Household) UpdatedAt() time.Time       { return h.updatedAt }
