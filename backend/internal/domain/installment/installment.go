package installment

import (
	"strings"
	"time"

	"micha/backend/internal/domain/shared"
)

// ID is the unique identifier type for an installment.
type ID string

// InstallmentAttributes is the flat DTO used for construction and rehydration.
type InstallmentAttributes struct {
	ID                     ID
	ExpenseID              string
	PaidByMemberID         string    // Member who paid the root MSI expense
	StartDate              time.Time // Date when this installment is due (for period matching)
	InstallmentAmountCents int64     // Amount for this specific installment
	TotalAmountCents       int64
	TotalInstallments      int
	CurrentInstallment     int
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// Installment represents a monthly payment for an MSI (meses sin intereses) purchase.
// An installment lives in the period when it's charged, not when the purchase was made.
type Installment struct {
	id                     ID
	expenseID              string
	paidByMemberID         string    // Member who paid the root MSI expense
	startDate              time.Time // Date when this installment is due
	installmentAmountCents int64     // Amount for this specific installment
	totalAmountCents       int64
	totalInstallments      int
	currentInstallment     int
	createdAt              time.Time
	updatedAt              time.Time
}

// New constructs an Installment from individual fields.
func New(id ID, expenseID string, paidByMemberID string, startDate time.Time, installmentAmountCents int64, totalAmountCents int64, totalInstallments int, currentInstallment int, createdAt time.Time) (Installment, error) {
	return NewFromAttributes(InstallmentAttributes{
		ID:                     id,
		ExpenseID:              expenseID,
		PaidByMemberID:         paidByMemberID,
		StartDate:              startDate,
		InstallmentAmountCents: installmentAmountCents,
		TotalAmountCents:       totalAmountCents,
		TotalInstallments:      totalInstallments,
		CurrentInstallment:     currentInstallment,
		CreatedAt:              createdAt,
		UpdatedAt:              createdAt,
	})
}

// NewFromAttributes constructs an Installment from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs InstallmentAttributes) (Installment, error) {
	if strings.TrimSpace(string(attrs.ID)) == "" {
		return Installment{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.ExpenseID) == "" {
		return Installment{}, shared.ErrInvalidID
	}

	if strings.TrimSpace(attrs.PaidByMemberID) == "" {
		return Installment{}, shared.ErrInvalidID
	}

	if attrs.InstallmentAmountCents <= 0 {
		return Installment{}, shared.ErrInvalidMoney
	}

	if attrs.TotalAmountCents <= 0 {
		return Installment{}, shared.ErrInvalidMoney
	}

	if attrs.TotalInstallments <= 0 {
		return Installment{}, shared.ErrInvalidStatus
	}

	if attrs.CurrentInstallment < 1 || attrs.CurrentInstallment > attrs.TotalInstallments {
		return Installment{}, shared.ErrInvalidStatus
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Installment{
		id:                     attrs.ID,
		expenseID:              attrs.ExpenseID,
		paidByMemberID:         attrs.PaidByMemberID,
		startDate:              attrs.StartDate,
		installmentAmountCents: attrs.InstallmentAmountCents,
		totalAmountCents:       attrs.TotalAmountCents,
		totalInstallments:      attrs.TotalInstallments,
		currentInstallment:     attrs.CurrentInstallment,
		createdAt:              attrs.CreatedAt,
		updatedAt:              updatedAt,
	}, nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (i Installment) Attributes() InstallmentAttributes {
	return InstallmentAttributes{
		ID:                     i.id,
		ExpenseID:              i.expenseID,
		PaidByMemberID:         i.paidByMemberID,
		StartDate:              i.startDate,
		InstallmentAmountCents: i.installmentAmountCents,
		TotalAmountCents:       i.totalAmountCents,
		TotalInstallments:      i.totalInstallments,
		CurrentInstallment:     i.currentInstallment,
		CreatedAt:              i.createdAt,
		UpdatedAt:              i.updatedAt,
	}
}

func (i Installment) ID() ID                        { return i.id }
func (i Installment) ExpenseID() string             { return i.expenseID }
func (i Installment) PaidByMemberID() string        { return i.paidByMemberID }
func (i Installment) StartDate() time.Time          { return i.startDate }
func (i Installment) InstallmentAmountCents() int64 { return i.installmentAmountCents }
func (i Installment) TotalAmountCents() int64       { return i.totalAmountCents }
func (i Installment) TotalInstallments() int        { return i.totalInstallments }
func (i Installment) CurrentInstallment() int       { return i.currentInstallment }
func (i Installment) CreatedAt() time.Time          { return i.createdAt }
func (i Installment) UpdatedAt() time.Time          { return i.updatedAt }
