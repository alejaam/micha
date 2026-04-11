package invitation

import (
	"errors"
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

var (
	ErrInvalidHouseholdID = errors.New("invalid household id")
	ErrInvalidMemberID    = errors.New("invalid member id")
	ErrInvalidEmail       = errors.New("invalid invitation email")
	ErrInvalidCode        = errors.New("invalid invitation code")
)

// ID is the unique identifier type for an invitation.
type ID string

// Attributes is the flat DTO used for construction and rehydration.
type Attributes struct {
	ID          ID
	HouseholdID string
	MemberID    string
	Email       string
	Code        string
	ExpiresAt   time.Time
	UsedAt      *time.Time
	CreatedAt   time.Time
}

// Invitation represents an invitation code sent to a future member.
type Invitation struct {
	id          ID
	householdID string
	memberID    string
	email       string
	code        string
	expiresAt   time.Time
	usedAt      *time.Time
	createdAt   time.Time
}

// NewFromAttributes builds an invitation entity from a flat attribute bag.
func NewFromAttributes(attrs Attributes) (Invitation, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Invitation{}, shared.ErrInvalidID
	}
	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return Invitation{}, ErrInvalidHouseholdID
	}
	if strings.TrimSpace(attrs.MemberID) == "" {
		return Invitation{}, ErrInvalidMemberID
	}
	email := strings.ToLower(strings.TrimSpace(attrs.Email))
	if email == "" {
		return Invitation{}, ErrInvalidEmail
	}
	code := strings.TrimSpace(attrs.Code)
	if code == "" {
		return Invitation{}, ErrInvalidCode
	}
	if attrs.ExpiresAt.IsZero() {
		return Invitation{}, ErrInvalidCode
	}

	return Invitation{
		id:          attrs.ID,
		householdID: attrs.HouseholdID,
		memberID:    attrs.MemberID,
		email:       email,
		code:        code,
		expiresAt:   attrs.ExpiresAt,
		usedAt:      attrs.UsedAt,
		createdAt:   attrs.CreatedAt,
	}, nil
}

func (i Invitation) Attributes() Attributes {
	return Attributes{
		ID:          i.id,
		HouseholdID: i.householdID,
		MemberID:    i.memberID,
		Email:       i.email,
		Code:        i.code,
		ExpiresAt:   i.expiresAt,
		UsedAt:      i.usedAt,
		CreatedAt:   i.createdAt,
	}
}

func (i Invitation) ID() ID               { return i.id }
func (i Invitation) HouseholdID() string  { return i.householdID }
func (i Invitation) MemberID() string     { return i.memberID }
func (i Invitation) Email() string        { return i.email }
func (i Invitation) Code() string         { return i.code }
func (i Invitation) ExpiresAt() time.Time { return i.expiresAt }
func (i Invitation) UsedAt() *time.Time   { return i.usedAt }
func (i Invitation) CreatedAt() time.Time { return i.createdAt }
