package member

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidName   = errors.New("invalid member name")
	ErrInvalidEmail  = errors.New("invalid member email")
	ErrInvalidSalary = errors.New("invalid member salary")
)

// ID is the unique identifier type for a member.
type ID string

// Attributes is the flat DTO used for construction and rehydration.
type Attributes struct {
	ID                 ID
	HouseholdID        string
	Name               string
	Email              string
	MonthlySalaryCents int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Member is the aggregate root for household members.
type Member struct {
	id                 ID
	householdID        string
	name               string
	email              string
	monthlySalaryCents int64
	createdAt          time.Time
	updatedAt          time.Time
}

// New constructs a Member from individual fields.
func New(id ID, householdID, name, email string, monthlySalaryCents int64, createdAt time.Time) (Member, error) {
	return NewFromAttributes(Attributes{
		ID:                 id,
		HouseholdID:        householdID,
		Name:               name,
		Email:              email,
		MonthlySalaryCents: monthlySalaryCents,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	})
}

// NewFromAttributes constructs a Member from a flat attribute bag.
func NewFromAttributes(attrs Attributes) (Member, error) {
	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Member{}, ErrInvalidName
	}

	email := strings.TrimSpace(attrs.Email)
	if !strings.Contains(email, "@") {
		return Member{}, ErrInvalidEmail
	}

	if attrs.MonthlySalaryCents < 0 {
		return Member{}, ErrInvalidSalary
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Member{
		id:                 attrs.ID,
		householdID:        attrs.HouseholdID,
		name:               name,
		email:              strings.ToLower(email),
		monthlySalaryCents: attrs.MonthlySalaryCents,
		createdAt:          attrs.CreatedAt,
		updatedAt:          updatedAt,
	}, nil
}

// UpdateProfile updates mutable member fields.
func (m *Member) UpdateProfile(name, email string, monthlySalaryCents int64) error {
	candidate, err := NewFromAttributes(Attributes{
		ID:                 m.id,
		HouseholdID:        m.householdID,
		Name:               name,
		Email:              email,
		MonthlySalaryCents: monthlySalaryCents,
		CreatedAt:          m.createdAt,
		UpdatedAt:          time.Now(),
	})
	if err != nil {
		return err
	}

	m.name = candidate.name
	m.email = candidate.email
	m.monthlySalaryCents = candidate.monthlySalaryCents
	m.updatedAt = candidate.updatedAt
	return nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (m Member) Attributes() Attributes {
	return Attributes{
		ID:                 m.id,
		HouseholdID:        m.householdID,
		Name:               m.name,
		Email:              m.email,
		MonthlySalaryCents: m.monthlySalaryCents,
		CreatedAt:          m.createdAt,
		UpdatedAt:          m.updatedAt,
	}
}

func (m Member) ID() ID                    { return m.id }
func (m Member) HouseholdID() string       { return m.householdID }
func (m Member) Name() string              { return m.name }
func (m Member) Email() string             { return m.email }
func (m Member) MonthlySalaryCents() int64 { return m.monthlySalaryCents }
func (m Member) CreatedAt() time.Time      { return m.createdAt }
func (m Member) UpdatedAt() time.Time      { return m.updatedAt }
