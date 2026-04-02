package member

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

var (
	ErrInvalidName   = errors.New("invalid member name")
	ErrInvalidEmail  = errors.New("invalid member email")
	ErrInvalidSalary = errors.New("invalid member monthly salary")
)

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ID is the unique identifier type for a member.
type ID string

// Attributes is the flat DTO used for construction and rehydration.
type Attributes struct {
	ID                 ID
	HouseholdID        string
	Name               string
	Email              string
	UserID             string // foreign key to users; may be empty for pending invites
	MonthlySalaryCents int64  // in cents; 0 is valid (no salary)
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Member represents a person within a Household.
// MonthlySalaryCents defines their contribution basis for proportional expense splits.
type Member struct {
	id                 ID
	householdID        string
	name               string
	email              string
	userID             string
	monthlySalaryCents int64
	createdAt          time.Time
	updatedAt          time.Time
}

// New constructs a Member from individual fields without a linked user.
func New(id ID, householdID string, name string, email string, monthlySalaryCents int64, createdAt time.Time) (Member, error) {
	return NewFromAttributes(Attributes{
		ID:                 id,
		HouseholdID:        householdID,
		Name:               name,
		Email:              email,
		UserID:             "",
		MonthlySalaryCents: monthlySalaryCents,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	})
}

// NewWithUserID constructs a Member with a linked user.
func NewWithUserID(id ID, householdID string, name string, email string, userID string, monthlySalaryCents int64, createdAt time.Time) (Member, error) {
	return NewFromAttributes(Attributes{
		ID:                 id,
		HouseholdID:        householdID,
		Name:               name,
		Email:              email,
		UserID:             userID,
		MonthlySalaryCents: monthlySalaryCents,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	})
}

// NewFromAttributes constructs a Member from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs Attributes) (Member, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Member{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return Member{}, shared.ErrInvalidID
	}

	name := strings.TrimSpace(attrs.Name)
	if name == "" {
		return Member{}, ErrInvalidName
	}

	email := strings.ToLower(strings.TrimSpace(attrs.Email))
	if !emailPattern.MatchString(email) {
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
		email:              email,
		userID:             strings.TrimSpace(attrs.UserID),
		monthlySalaryCents: attrs.MonthlySalaryCents,
		createdAt:          attrs.CreatedAt,
		updatedAt:          updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (m Member) Attributes() Attributes {
	return Attributes{
		ID:                 m.id,
		HouseholdID:        m.householdID,
		Name:               m.name,
		Email:              m.email,
		UserID:             m.userID,
		MonthlySalaryCents: m.monthlySalaryCents,
		CreatedAt:          m.createdAt,
		UpdatedAt:          m.updatedAt,
	}
}

// LinkUser associates this member with a user account.
func (m *Member) LinkUser(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return shared.ErrInvalidID
	}
	m.userID = userID
	m.updatedAt = time.Now()
	return nil
}

// UpdateProfile updates member personal information.
func (m *Member) UpdateProfile(name string, email string, monthlySalaryCents int64) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !emailPattern.MatchString(email) {
		return ErrInvalidEmail
	}

	if monthlySalaryCents < 0 {
		return ErrInvalidSalary
	}

	m.name = name
	m.email = email
	m.monthlySalaryCents = monthlySalaryCents
	m.updatedAt = time.Now()

	return nil
}

func (m Member) ID() ID                    { return m.id }
func (m Member) HouseholdID() string       { return m.householdID }
func (m Member) Name() string              { return m.name }
func (m Member) Email() string             { return m.email }
func (m Member) UserID() string            { return m.userID }
func (m Member) IsPending() bool           { return strings.TrimSpace(m.userID) == "" }
func (m Member) MonthlySalaryCents() int64 { return m.monthlySalaryCents }
func (m Member) CreatedAt() time.Time      { return m.createdAt }
func (m Member) UpdatedAt() time.Time      { return m.updatedAt }
