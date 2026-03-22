package income

import (
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// ID is the unique identifier type for an income.
type ID string

// IncomeAttributes is the flat DTO used for construction and rehydration.
type IncomeAttributes struct {
	ID          ID
	MemberID    string
	PeriodID    string
	AmountCents int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Income represents the declared income of a member for a specific period.
// Used to calculate suggested contribution percentages.
type Income struct {
	id          ID
	memberID    string
	periodID    string
	amountCents int64
	createdAt   time.Time
	updatedAt   time.Time
}

// New constructs an Income from individual fields.
func New(id ID, memberID string, periodID string, amountCents int64, createdAt time.Time) (Income, error) {
	return NewFromAttributes(IncomeAttributes{
		ID:          id,
		MemberID:    memberID,
		PeriodID:    periodID,
		AmountCents: amountCents,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	})
}

// NewFromAttributes constructs an Income from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs IncomeAttributes) (Income, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Income{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.MemberID) == "" {
		return Income{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.PeriodID) == "" {
		return Income{}, shared.ErrInvalidID
	}

	if attrs.AmountCents <= 0 {
		return Income{}, shared.ErrInvalidMoney
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Income{
		id:          attrs.ID,
		memberID:    attrs.MemberID,
		periodID:    attrs.PeriodID,
		amountCents: attrs.AmountCents,
		createdAt:   attrs.CreatedAt,
		updatedAt:   updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (i Income) Attributes() IncomeAttributes {
	return IncomeAttributes{
		ID:          i.id,
		MemberID:    i.memberID,
		PeriodID:    i.periodID,
		AmountCents: i.amountCents,
		CreatedAt:   i.createdAt,
		UpdatedAt:   i.updatedAt,
	}
}

func (i Income) ID() ID               { return i.id }
func (i Income) MemberID() string     { return i.memberID }
func (i Income) PeriodID() string     { return i.periodID }
func (i Income) AmountCents() int64   { return i.amountCents }
func (i Income) CreatedAt() time.Time { return i.createdAt }
func (i Income) UpdatedAt() time.Time { return i.updatedAt }
