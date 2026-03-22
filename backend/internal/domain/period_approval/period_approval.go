package periodapproval

import (
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// ApprovalStatus defines whether a member approved or objected to period closure.
type ApprovalStatus string

const (
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusObjected ApprovalStatus = "objected"
)

// ID is the unique identifier type for a period approval.
type ID string

// PeriodApprovalAttributes is the flat DTO used for construction and rehydration.
type PeriodApprovalAttributes struct {
	ID        ID
	MemberID  string
	PeriodID  string
	Status    ApprovalStatus
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PeriodApproval represents a member's approval or objection to closing a period.
// Used to enforce consensus before period closure.
type PeriodApproval struct {
	id        ID
	memberID  string
	periodID  string
	status    ApprovalStatus
	comment   string
	createdAt time.Time
	updatedAt time.Time
}

// New constructs a PeriodApproval from individual fields.
func New(id ID, memberID string, periodID string, status ApprovalStatus, comment string, createdAt time.Time) (PeriodApproval, error) {
	return NewFromAttributes(PeriodApprovalAttributes{
		ID:        id,
		MemberID:  memberID,
		PeriodID:  periodID,
		Status:    status,
		Comment:   comment,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
}

// NewFromAttributes constructs a PeriodApproval from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs PeriodApprovalAttributes) (PeriodApproval, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return PeriodApproval{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.MemberID) == "" {
		return PeriodApproval{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.PeriodID) == "" {
		return PeriodApproval{}, shared.ErrInvalidID
	}

	status := attrs.Status
	if status == "" {
		status = ApprovalStatusApproved
	}
	if status != ApprovalStatusApproved && status != ApprovalStatusObjected {
		return PeriodApproval{}, shared.ErrInvalidStatus
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return PeriodApproval{
		id:        attrs.ID,
		memberID:  attrs.MemberID,
		periodID:  attrs.PeriodID,
		status:    status,
		comment:   strings.TrimSpace(attrs.Comment),
		createdAt: attrs.CreatedAt,
		updatedAt: updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (p PeriodApproval) Attributes() PeriodApprovalAttributes {
	return PeriodApprovalAttributes{
		ID:        p.id,
		MemberID:  p.memberID,
		PeriodID:  p.periodID,
		Status:    p.status,
		Comment:   p.comment,
		CreatedAt: p.createdAt,
		UpdatedAt: p.updatedAt,
	}
}

func (p PeriodApproval) ID() ID                 { return p.id }
func (p PeriodApproval) MemberID() string       { return p.memberID }
func (p PeriodApproval) PeriodID() string       { return p.periodID }
func (p PeriodApproval) Status() ApprovalStatus { return p.status }
func (p PeriodApproval) Comment() string        { return p.comment }
func (p PeriodApproval) CreatedAt() time.Time   { return p.createdAt }
func (p PeriodApproval) UpdatedAt() time.Time   { return p.updatedAt }
