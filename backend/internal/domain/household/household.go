package household

import (
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// CycleType defines the period duration for a household.
type CycleType string

const (
	CycleTypeMonthly  CycleType = "monthly"
	CycleTypeBiweekly CycleType = "biweekly"
	CycleTypeCustom   CycleType = "custom"
)

// ID is the unique identifier type for a household.
type ID string

// HouseholdAttributes is the flat DTO used for construction and rehydration.
type HouseholdAttributes struct {
	ID        ID
	Name      string
	OwnerID   string
	CycleType CycleType
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Household is the aggregate root for a household.
type Household struct {
	id        ID
	name      string
	ownerID   string
	cycleType CycleType
	createdAt time.Time
	updatedAt time.Time
}

// New constructs a Household from individual fields.
func New(id ID, name string, ownerID string, cycleType CycleType, createdAt time.Time) (Household, error) {
	return NewFromAttributes(HouseholdAttributes{
		ID:        id,
		Name:      name,
		OwnerID:   ownerID,
		CycleType: cycleType,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
}

// NewFromAttributes constructs a Household from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs HouseholdAttributes) (Household, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Household{}, shared.ErrInvalidID
	}

	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Household{}, shared.ErrInvalidName
	}

	if strings.TrimSpace(attrs.OwnerID) == "" {
		return Household{}, shared.ErrInvalidID
	}

	cycleType := attrs.CycleType
	if cycleType == "" {
		cycleType = CycleTypeMonthly
	}
	if cycleType != CycleTypeMonthly && cycleType != CycleTypeBiweekly && cycleType != CycleTypeCustom {
		return Household{}, shared.ErrInvalidStatus
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Household{
		id:        attrs.ID,
		name:      name,
		ownerID:   attrs.OwnerID,
		cycleType: cycleType,
		createdAt: attrs.CreatedAt,
		updatedAt: updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (h Household) Attributes() HouseholdAttributes {
	return HouseholdAttributes{
		ID:        h.id,
		Name:      h.name,
		OwnerID:   h.ownerID,
		CycleType: h.cycleType,
		CreatedAt: h.createdAt,
		UpdatedAt: h.updatedAt,
	}
}

func (h Household) ID() ID               { return h.id }
func (h Household) Name() string         { return h.name }
func (h Household) OwnerID() string      { return h.ownerID }
func (h Household) CycleType() CycleType { return h.cycleType }
func (h Household) CreatedAt() time.Time { return h.createdAt }
func (h Household) UpdatedAt() time.Time { return h.updatedAt }
