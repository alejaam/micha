package expense

import (
	"time"

	"micha/backend/internal/domain/shared"
)

// ID is the unique identifier type for an expense.
type ID string

// ExpenseAttributes is the flat DTO used for construction and rehydration.
type ExpenseAttributes struct {
	ID          ID
	HouseholdID string
	AmountCents int64
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Expense is the aggregate root for an expense record.
type Expense struct {
	id          ID
	householdID string
	amountCents int64
	description string
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// New constructs an Expense from individual fields.
func New(id ID, householdID string, amountCents int64, description string, createdAt time.Time) (Expense, error) {
	return NewFromAttributes(ExpenseAttributes{
		ID:          id,
		HouseholdID: householdID,
		AmountCents: amountCents,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	})
}

// NewFromAttributes constructs an Expense from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs ExpenseAttributes) (Expense, error) {
	if attrs.AmountCents <= 0 {
		return Expense{}, shared.ErrInvalidMoney
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Expense{
		id:          attrs.ID,
		householdID: attrs.HouseholdID,
		amountCents: attrs.AmountCents,
		description: attrs.Description,
		createdAt:   attrs.CreatedAt,
		updatedAt:   updatedAt,
		deletedAt:   attrs.DeletedAt,
	}, nil
}

// Patch applies a partial update to the expense.
// Only non-nil fields are modified. Invariants are re-validated after the change.
func (e *Expense) Patch(description *string, amountCents *int64) error {
	if amountCents != nil {
		if *amountCents <= 0 {
			return shared.ErrInvalidMoney
		}
		e.amountCents = *amountCents
	}
	if description != nil {
		e.description = *description
	}
	e.updatedAt = time.Now()
	return nil
}

// SoftDelete marks the expense as deleted without removing it from the store.
func (e *Expense) SoftDelete() error {
	if e.deletedAt != nil {
		return shared.ErrAlreadyDeleted
	}
	now := time.Now()
	e.deletedAt = &now
	e.updatedAt = now
	return nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (e Expense) Attributes() ExpenseAttributes {
	return ExpenseAttributes{
		ID:          e.id,
		HouseholdID: e.householdID,
		AmountCents: e.amountCents,
		Description: e.description,
		CreatedAt:   e.createdAt,
		UpdatedAt:   e.updatedAt,
		DeletedAt:   e.deletedAt,
	}
}

func (e Expense) ID() ID                { return e.id }
func (e Expense) HouseholdID() string   { return e.householdID }
func (e Expense) AmountCents() int64    { return e.amountCents }
func (e Expense) Description() string   { return e.description }
func (e Expense) CreatedAt() time.Time  { return e.createdAt }
func (e Expense) UpdatedAt() time.Time  { return e.updatedAt }
func (e Expense) DeletedAt() *time.Time { return e.deletedAt }
