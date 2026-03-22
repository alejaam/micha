package period

import (
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// Status defines the lifecycle state of a period.
type Status string

const (
	StatusOpen   Status = "open"
	StatusReview Status = "review"
	StatusClosed Status = "closed"
)

// ID is the unique identifier type for a period.
type ID string

// PeriodAttributes is the flat DTO used for construction and rehydration.
type PeriodAttributes struct {
	ID          ID
	HouseholdID string
	StartDate   time.Time
	EndDate     time.Time
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Period represents a financial cycle for a household.
// Only one period with status=open can exist per household at a time.
// During status=review, no new expenses can be added.
type Period struct {
	id          ID
	householdID string
	startDate   time.Time
	endDate     time.Time
	status      Status
	createdAt   time.Time
	updatedAt   time.Time
}

// New constructs a Period from individual fields.
func New(id ID, householdID string, startDate time.Time, endDate time.Time, status Status, createdAt time.Time) (Period, error) {
	return NewFromAttributes(PeriodAttributes{
		ID:          id,
		HouseholdID: householdID,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      status,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	})
}

// NewFromAttributes constructs a Period from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs PeriodAttributes) (Period, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Period{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return Period{}, shared.ErrInvalidID
	}

	if attrs.StartDate.After(attrs.EndDate) {
		return Period{}, shared.ErrInvalidDateRange
	}

	status := attrs.Status
	if status == "" {
		status = StatusOpen
	}
	if status != StatusOpen && status != StatusReview && status != StatusClosed {
		return Period{}, shared.ErrInvalidStatus
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Period{
		id:          attrs.ID,
		householdID: attrs.HouseholdID,
		startDate:   attrs.StartDate,
		endDate:     attrs.EndDate,
		status:      status,
		createdAt:   attrs.CreatedAt,
		updatedAt:   updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (p Period) Attributes() PeriodAttributes {
	return PeriodAttributes{
		ID:          p.id,
		HouseholdID: p.householdID,
		StartDate:   p.startDate,
		EndDate:     p.endDate,
		Status:      p.status,
		CreatedAt:   p.createdAt,
		UpdatedAt:   p.updatedAt,
	}
}

func (p Period) ID() ID               { return p.id }
func (p Period) HouseholdID() string  { return p.householdID }
func (p Period) StartDate() time.Time { return p.startDate }
func (p Period) EndDate() time.Time   { return p.endDate }
func (p Period) Status() Status       { return p.status }
func (p Period) CreatedAt() time.Time { return p.createdAt }
func (p Period) UpdatedAt() time.Time { return p.updatedAt }
