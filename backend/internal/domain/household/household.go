package household

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

var (
	ErrInvalidName           = errors.New("invalid household name")
	ErrInvalidSettlementMode = errors.New("invalid settlement mode")
	ErrInvalidCurrency       = errors.New("invalid currency")
	ErrInvalidSplitConfig    = errors.New("invalid split config")
	ErrEmptySplitConfig      = errors.New("empty split config")
)

// SettlementMode defines how expenses are divided among members.
type SettlementMode string

const (
	SettlementModeEqual        SettlementMode = "equal"
	SettlementModeProportional SettlementMode = "proportional"
)

// ID is the unique identifier type for a household.
type ID string

// Attributes is the flat DTO used for construction and rehydration.
type Attributes struct {
	ID             ID
	Name           string
	OwnerID        string
	SettlementMode SettlementMode
	Currency       string // ISO 4217 code (uppercase)
	SplitConfig    SplitConfig
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Household is the aggregate root for a household.
type Household struct {
	id             ID
	name           string
	ownerID        string
	settlementMode SettlementMode
	currency       string
	createdAt      time.Time
	updatedAt      time.Time
}

// New constructs a Household from individual fields.
func New(id ID, name string, ownerID string, settlementMode SettlementMode, currency string, createdAt time.Time) (Household, error) {
	return NewFromAttributes(Attributes{
		ID:             id,
		Name:           name,
		OwnerID:        ownerID,
		SettlementMode: settlementMode,
		Currency:       currency,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})
}

// NewFromAttributes constructs a Household from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs Attributes) (Household, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Household{}, shared.ErrInvalidID
	}

	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Household{}, ErrInvalidName
	}

	// Validate SettlementMode
	if attrs.SettlementMode != SettlementModeEqual && attrs.SettlementMode != SettlementModeProportional {
		return Household{}, ErrInvalidSettlementMode
	}

	// Validate and normalize currency to uppercase ISO 4217
	currency := strings.ToUpper(strings.TrimSpace(attrs.Currency))
	if !isValidISO4217(currency) {
		return Household{}, ErrInvalidCurrency
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Household{
		id:             attrs.ID,
		name:           name,
		ownerID:        strings.TrimSpace(attrs.OwnerID),
		settlementMode: attrs.SettlementMode,
		currency:       currency,
		createdAt:      attrs.CreatedAt,
		updatedAt:      updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (h Household) Attributes() Attributes {
	return Attributes{
		ID:             h.id,
		Name:           h.name,
		OwnerID:        h.ownerID,
		SettlementMode: h.settlementMode,
		Currency:       h.currency,
		SplitConfig:    SplitConfig{}, // Always empty for now; loaded separately by adapters
		CreatedAt:      h.createdAt,
		UpdatedAt:      h.updatedAt,
	}
}

// UpdateConfig updates the household name, settlement mode and currency.
func (h *Household) UpdateConfig(name string, settlementMode SettlementMode, currency string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}

	if settlementMode != SettlementModeEqual && settlementMode != SettlementModeProportional {
		return ErrInvalidSettlementMode
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if !isValidISO4217(currency) {
		return ErrInvalidCurrency
	}

	h.name = name
	h.settlementMode = settlementMode
	h.currency = currency
	h.updatedAt = time.Now()

	return nil
}

// UpdateSplitConfig is a placeholder method; actual SplitConfig management is handled by adapters.
// This method exists for compatibility with use case interfaces.
func (h *Household) UpdateSplitConfig(sc SplitConfig) {
	// SplitConfig is managed separately via adapters; this is a no-op in domain context.
	h.updatedAt = time.Now()
}

func (h Household) ID() ID                         { return h.id }
func (h Household) Name() string                   { return h.name }
func (h Household) OwnerID() string                { return h.ownerID }
func (h Household) SettlementMode() SettlementMode { return h.settlementMode }
func (h Household) Currency() string               { return h.currency }
func (h Household) CreatedAt() time.Time           { return h.createdAt }
func (h Household) UpdatedAt() time.Time           { return h.updatedAt }

// isValidISO4217 checks if a currency code is a valid 3-letter ISO 4217 code.
func isValidISO4217(code string) bool {
	// Simple validation: 3 uppercase letters
	matched, _ := regexp.MatchString(`^[A-Z]{3}$`, code)
	return matched
}
