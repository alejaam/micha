package expense

import (
	"time"

	"micha/backend/internal/domain/shared"
)

type ID string

type Expense struct {
	id          ID
	householdID string
	amountCents int64
	description string
	createdAt   time.Time
}

func New(id ID, householdID string, amountCents int64, description string, createdAt time.Time) (Expense, error) {
	if amountCents <= 0 {
		return Expense{}, shared.ErrInvalidMoney
	}

	return Expense{
		id:          id,
		householdID: householdID,
		amountCents: amountCents,
		description: description,
		createdAt:   createdAt,
	}, nil
}

func (e Expense) ID() ID               { return e.id }
func (e Expense) HouseholdID() string  { return e.householdID }
func (e Expense) AmountCents() int64   { return e.amountCents }
func (e Expense) Description() string  { return e.description }
func (e Expense) CreatedAt() time.Time { return e.createdAt }
