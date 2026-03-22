package member

import (
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// ID is the unique identifier type for a member.
type ID string

// MemberAttributes is the flat DTO used for construction and rehydration.
type MemberAttributes struct {
	ID              ID
	HouseholdID     string
	UserID          string
	ContributionPct float64
	ValidFrom       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Member represents a person within a Household.
// ContributionPct defines the proportion of shared expenses they pay.
// ValidFrom ensures the % only applies forward, never retroactively.
type Member struct {
	id              ID
	householdID     string
	userID          string
	contributionPct float64
	validFrom       time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// New constructs a Member from individual fields.
func New(id ID, householdID string, userID string, contributionPct float64, validFrom time.Time, createdAt time.Time) (Member, error) {
	return NewFromAttributes(MemberAttributes{
		ID:              id,
		HouseholdID:     householdID,
		UserID:          userID,
		ContributionPct: contributionPct,
		ValidFrom:       validFrom,
		CreatedAt:       createdAt,
		UpdatedAt:       createdAt,
	})
}

// NewFromAttributes constructs a Member from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs MemberAttributes) (Member, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Member{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return Member{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.UserID) == "" {
		return Member{}, shared.ErrInvalidID
	}

	if attrs.ContributionPct < 0 || attrs.ContributionPct > 100 {
		return Member{}, shared.ErrInvalidPercentage
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Member{
		id:              attrs.ID,
		householdID:     attrs.HouseholdID,
		userID:          attrs.UserID,
		contributionPct: attrs.ContributionPct,
		validFrom:       attrs.ValidFrom,
		createdAt:       attrs.CreatedAt,
		updatedAt:       updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (m Member) Attributes() MemberAttributes {
	return MemberAttributes{
		ID:              m.id,
		HouseholdID:     m.householdID,
		UserID:          m.userID,
		ContributionPct: m.contributionPct,
		ValidFrom:       m.validFrom,
		CreatedAt:       m.createdAt,
		UpdatedAt:       m.updatedAt,
	}
}

func (m Member) ID() ID                   { return m.id }
func (m Member) HouseholdID() string      { return m.householdID }
func (m Member) UserID() string           { return m.userID }
func (m Member) ContributionPct() float64 { return m.contributionPct }
func (m Member) ValidFrom() time.Time     { return m.validFrom }
func (m Member) CreatedAt() time.Time     { return m.createdAt }
func (m Member) UpdatedAt() time.Time     { return m.updatedAt }
