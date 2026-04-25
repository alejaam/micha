package card

import (
	"errors"
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

var (
	ErrInvalidBankName  = errors.New("invalid bank name")
	ErrInvalidCardName  = errors.New("invalid card name")
	ErrInvalidCutoffDay = errors.New("invalid cutoff day: must be between 1 and 31")
)

// ID is the unique identifier type for a card.
type ID string

// Attributes is the flat DTO used for construction and rehydration.
type Attributes struct {
	ID            ID
	HouseholdID   string
	OwnerMemberID string
	BankName      string // e.g., "BANAMEX", "BBVA", "Nu", "HSBC", "Rappi"
	CardName      string // User-friendly label, e.g., "Banamex Oro", "BBVA Azul"
	CutoffDay     int    // Day of month (1-31) when the statement closes
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// Card represents a credit card used for shared expenses.
// Each card belongs to a household and tracks the billing cutoff day.
type Card struct {
	id            ID
	householdID   string
	ownerMemberID string
	bankName      string
	cardName      string
	cutoffDay     int
	createdAt     time.Time
	updatedAt     time.Time
	deletedAt     *time.Time
}

// New constructs a Card from individual fields.
func New(id ID, householdID string, ownerMemberID string, bankName string, cardName string, cutoffDay int, createdAt time.Time) (Card, error) {
	return NewFromAttributes(Attributes{
		ID:            id,
		HouseholdID:   householdID,
		OwnerMemberID: ownerMemberID,
		BankName:      bankName,
		CardName:      cardName,
		CutoffDay:     cutoffDay,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
		DeletedAt:     nil,
	})
}

// NewFromAttributes constructs a Card from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs Attributes) (Card, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Card{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return Card{}, shared.ErrInvalidID
	}

	bankName := strings.TrimSpace(attrs.BankName)
	if bankName == "" {
		return Card{}, ErrInvalidBankName
	}

	cardName := strings.TrimSpace(attrs.CardName)
	if cardName == "" {
		return Card{}, ErrInvalidCardName
	}

	if attrs.CutoffDay < 1 || attrs.CutoffDay > 31 {
		return Card{}, ErrInvalidCutoffDay
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Card{
		id:            attrs.ID,
		householdID:   attrs.HouseholdID,
		ownerMemberID: strings.TrimSpace(attrs.OwnerMemberID),
		bankName:      bankName,
		cardName:      cardName,
		cutoffDay:     attrs.CutoffDay,
		createdAt:     attrs.CreatedAt,
		updatedAt:     updatedAt,
		deletedAt:     attrs.DeletedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (c Card) Attributes() Attributes {
	return Attributes{
		ID:            c.id,
		HouseholdID:   c.householdID,
		OwnerMemberID: c.ownerMemberID,
		BankName:      c.bankName,
		CardName:      c.cardName,
		CutoffDay:     c.cutoffDay,
		CreatedAt:     c.createdAt,
		UpdatedAt:     c.updatedAt,
		DeletedAt:     c.deletedAt,
	}
}

// Getters
func (c Card) ID() ID                { return c.id }
func (c Card) HouseholdID() string   { return c.householdID }
func (c Card) OwnerMemberID() string { return c.ownerMemberID }
func (c Card) BankName() string      { return c.bankName }
func (c Card) CardName() string      { return c.cardName }
func (c Card) CutoffDay() int        { return c.cutoffDay }
func (c Card) CreatedAt() time.Time  { return c.createdAt }
func (c Card) UpdatedAt() time.Time  { return c.updatedAt }
func (c Card) DeletedAt() *time.Time { return c.deletedAt }

// IsOwnedBy returns true when the card has an owner and matches memberID.
func (c Card) IsOwnedBy(memberID string) bool {
	if c.ownerMemberID == "" {
		return false
	}
	return strings.TrimSpace(memberID) == c.ownerMemberID
}

// IsDeleted returns true if the card has been soft-deleted.
func (c Card) IsDeleted() bool {
	return c.deletedAt != nil
}

// Delete marks the card as deleted (soft delete).
func (c Card) Delete(deletedAt time.Time) Card {
	c.deletedAt = &deletedAt
	c.updatedAt = deletedAt
	return c
}
