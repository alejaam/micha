package expense

import (
	"errors"
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

var (
	ErrInvalidHouseholdID    = errors.New("invalid household id")
	ErrInvalidPaidByMemberID = errors.New("invalid paid by member id")
	ErrInvalidCurrency       = errors.New("invalid expense currency")
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
)

// PaymentMethod defines how an expense was paid.
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodVoucher  PaymentMethod = "voucher"
)

// ID is the unique identifier type for an expense.
type ID string

// ExpenseAttributes is the flat DTO used for construction and rehydration.
type ExpenseAttributes struct {
	ID             ID
	HouseholdID    string
	PaidByMemberID string
	AmountCents    int64
	Description    string
	IsShared       bool
	Currency       string
	PaymentMethod  PaymentMethod
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// Expense is the aggregate root for an expense record.
type Expense struct {
	id             ID
	householdID    string
	paidByMemberID string
	amountCents    int64
	description    string
	isShared       bool
	currency       string
	paymentMethod  PaymentMethod
	createdAt      time.Time
	updatedAt      time.Time
	deletedAt      *time.Time
}

// New constructs an Expense from individual fields.
func New(id ID, householdID string, amountCents int64, description string, createdAt time.Time) (Expense, error) {
	return NewFromAttributes(ExpenseAttributes{
		ID:             id,
		HouseholdID:    householdID,
		PaidByMemberID: "unassigned",
		AmountCents:    amountCents,
		Description:    description,
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  PaymentMethodCash,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})
}

// NewFromAttributes constructs an Expense from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs ExpenseAttributes) (Expense, error) {
	if attrs.AmountCents <= 0 {
		return Expense{}, shared.ErrInvalidMoney
	}

	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return Expense{}, ErrInvalidHouseholdID
	}

	paidByMemberID := strings.TrimSpace(attrs.PaidByMemberID)
	if paidByMemberID == "" {
		return Expense{}, ErrInvalidPaidByMemberID
	}

	currency := strings.ToUpper(strings.TrimSpace(attrs.Currency))
	if len(currency) != 3 {
		return Expense{}, ErrInvalidCurrency
	}

	paymentMethod := attrs.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = PaymentMethodCash
	}
	if paymentMethod != PaymentMethodCash && paymentMethod != PaymentMethodCard && paymentMethod != PaymentMethodTransfer && paymentMethod != PaymentMethodVoucher {
		return Expense{}, ErrInvalidPaymentMethod
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Expense{
		id:             attrs.ID,
		householdID:    attrs.HouseholdID,
		paidByMemberID: paidByMemberID,
		amountCents:    attrs.AmountCents,
		description:    attrs.Description,
		isShared:       attrs.IsShared,
		currency:       currency,
		paymentMethod:  paymentMethod,
		createdAt:      attrs.CreatedAt,
		updatedAt:      updatedAt,
		deletedAt:      attrs.DeletedAt,
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
		ID:             e.id,
		HouseholdID:    e.householdID,
		PaidByMemberID: e.paidByMemberID,
		AmountCents:    e.amountCents,
		Description:    e.description,
		IsShared:       e.isShared,
		Currency:       e.currency,
		PaymentMethod:  e.paymentMethod,
		CreatedAt:      e.createdAt,
		UpdatedAt:      e.updatedAt,
		DeletedAt:      e.deletedAt,
	}
}

func (e Expense) ID() ID                       { return e.id }
func (e Expense) HouseholdID() string          { return e.householdID }
func (e Expense) PaidByMemberID() string       { return e.paidByMemberID }
func (e Expense) AmountCents() int64           { return e.amountCents }
func (e Expense) Description() string          { return e.description }
func (e Expense) IsShared() bool               { return e.isShared }
func (e Expense) Currency() string             { return e.currency }
func (e Expense) PaymentMethod() PaymentMethod { return e.paymentMethod }
func (e Expense) CreatedAt() time.Time         { return e.createdAt }
func (e Expense) UpdatedAt() time.Time         { return e.updatedAt }
func (e Expense) DeletedAt() *time.Time        { return e.deletedAt }
