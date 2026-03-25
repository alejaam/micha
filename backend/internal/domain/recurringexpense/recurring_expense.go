package recurringexpense

import (
	"errors"
	"strings"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
)

var (
	ErrInvalidHouseholdID        = errors.New("invalid household id")
	ErrInvalidPaidByMemberID     = errors.New("invalid paid by member id")
	ErrInvalidRecurrencePattern  = errors.New("invalid recurrence pattern")
	ErrInvalidDateRange          = errors.New("end date must be after start date")
	ErrInvalidNextGenerationDate = errors.New("next generation date cannot be before start date")
)

// RecurrencePattern defines how often a recurring expense should be generated.
type RecurrencePattern string

const (
	RecurrencePatternMonthly  RecurrencePattern = "monthly"
	RecurrencePatternBiweekly RecurrencePattern = "biweekly"
	RecurrencePatternWeekly   RecurrencePattern = "weekly"
)

// ID is the unique identifier type for a recurring expense.
type ID string

// RecurringExpenseAttributes is the flat DTO used for construction and rehydration.
type RecurringExpenseAttributes struct {
	ID                 ID
	HouseholdID        string
	PaidByMemberID     string
	AmountCents        int64
	Description        string
	CategoryID         string
	ExpenseType        expense.ExpenseType
	RecurrencePattern  RecurrencePattern
	StartDate          time.Time
	EndDate            *time.Time
	NextGenerationDate time.Time
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// RecurringExpense is the aggregate root for a recurring expense template.
type RecurringExpense struct {
	id                 ID
	householdID        string
	paidByMemberID     string
	amountCents        int64
	description        string
	categoryID         string
	expenseType        expense.ExpenseType
	recurrencePattern  RecurrencePattern
	startDate          time.Time
	endDate            *time.Time
	nextGenerationDate time.Time
	isActive           bool
	createdAt          time.Time
	updatedAt          time.Time
	deletedAt          *time.Time
}

// New constructs a RecurringExpense from individual fields.
func New(
	id ID,
	householdID string,
	paidByMemberID string,
	amountCents int64,
	description string,
	categoryID string,
	expenseType expense.ExpenseType,
	recurrencePattern RecurrencePattern,
	startDate time.Time,
	createdAt time.Time,
) (RecurringExpense, error) {
	return NewFromAttributes(RecurringExpenseAttributes{
		ID:                 id,
		HouseholdID:        householdID,
		PaidByMemberID:     paidByMemberID,
		AmountCents:        amountCents,
		Description:        description,
		CategoryID:         categoryID,
		ExpenseType:        expenseType,
		RecurrencePattern:  recurrencePattern,
		StartDate:          startDate,
		EndDate:            nil,
		NextGenerationDate: startDate,
		IsActive:           true,
		CreatedAt:          createdAt,
		UpdatedAt:          createdAt,
	})
}

// NewFromAttributes constructs a RecurringExpense from a flat attribute bag (used for rehydration).
func NewFromAttributes(attrs RecurringExpenseAttributes) (RecurringExpense, error) {
	if attrs.AmountCents <= 0 {
		return RecurringExpense{}, shared.ErrInvalidMoney
	}

	if strings.TrimSpace(attrs.HouseholdID) == "" {
		return RecurringExpense{}, ErrInvalidHouseholdID
	}

	paidByMemberID := strings.TrimSpace(attrs.PaidByMemberID)
	if paidByMemberID == "" {
		return RecurringExpense{}, ErrInvalidPaidByMemberID
	}

	// Validate recurrence pattern
	if attrs.RecurrencePattern != RecurrencePatternMonthly &&
		attrs.RecurrencePattern != RecurrencePatternBiweekly &&
		attrs.RecurrencePattern != RecurrencePatternWeekly {
		return RecurringExpense{}, ErrInvalidRecurrencePattern
	}

	// Validate expense type
	if attrs.ExpenseType != expense.ExpenseTypeFixed &&
		attrs.ExpenseType != expense.ExpenseTypeVariable &&
		attrs.ExpenseType != expense.ExpenseTypeMSI {
		return RecurringExpense{}, expense.ErrInvalidExpenseType
	}

	// Validate date range
	if attrs.EndDate != nil && attrs.EndDate.Before(attrs.StartDate) {
		return RecurringExpense{}, ErrInvalidDateRange
	}

	// Validate next generation date
	if attrs.NextGenerationDate.Before(attrs.StartDate) {
		return RecurringExpense{}, ErrInvalidNextGenerationDate
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return RecurringExpense{
		id:                 attrs.ID,
		householdID:        attrs.HouseholdID,
		paidByMemberID:     paidByMemberID,
		amountCents:        attrs.AmountCents,
		description:        attrs.Description,
		categoryID:         strings.TrimSpace(attrs.CategoryID),
		expenseType:        attrs.ExpenseType,
		recurrencePattern:  attrs.RecurrencePattern,
		startDate:          attrs.StartDate,
		endDate:            attrs.EndDate,
		nextGenerationDate: attrs.NextGenerationDate,
		isActive:           attrs.IsActive,
		createdAt:          attrs.CreatedAt,
		updatedAt:          updatedAt,
		deletedAt:          attrs.DeletedAt,
	}, nil
}

// Patch applies a partial update to the recurring expense.
// Only non-nil fields are modified. Invariants are re-validated after the change.
func (r *RecurringExpense) Patch(
	description *string,
	amountCents *int64,
	categoryID *string,
	isActive *bool,
) error {
	if amountCents != nil {
		if *amountCents <= 0 {
			return shared.ErrInvalidMoney
		}
		r.amountCents = *amountCents
	}
	if description != nil {
		r.description = *description
	}
	if categoryID != nil {
		r.categoryID = strings.TrimSpace(*categoryID)
	}
	if isActive != nil {
		r.isActive = *isActive
	}
	r.updatedAt = time.Now()
	return nil
}

// AdvanceNextGenerationDate moves the next generation date forward based on the recurrence pattern.
func (r *RecurringExpense) AdvanceNextGenerationDate() {
	switch r.recurrencePattern {
	case RecurrencePatternMonthly:
		r.nextGenerationDate = r.nextGenerationDate.AddDate(0, 1, 0)
	case RecurrencePatternBiweekly:
		r.nextGenerationDate = r.nextGenerationDate.AddDate(0, 0, 14)
	case RecurrencePatternWeekly:
		r.nextGenerationDate = r.nextGenerationDate.AddDate(0, 0, 7)
	}
	r.updatedAt = time.Now()
}

// ShouldGenerateExpense returns true if an expense should be generated on the given date.
func (r RecurringExpense) ShouldGenerateExpense(now time.Time) bool {
	if !r.isActive {
		return false
	}
	if r.deletedAt != nil {
		return false
	}
	if now.Before(r.nextGenerationDate) {
		return false
	}
	if r.endDate != nil && now.After(*r.endDate) {
		return false
	}
	return true
}

// SoftDelete marks the recurring expense as deleted without removing it from the store.
func (r *RecurringExpense) SoftDelete() error {
	if r.deletedAt != nil {
		return shared.ErrAlreadyDeleted
	}
	now := time.Now()
	r.deletedAt = &now
	r.updatedAt = now
	return nil
}

// Attributes returns a copy of all fields as a flat DTO.
func (r RecurringExpense) Attributes() RecurringExpenseAttributes {
	return RecurringExpenseAttributes{
		ID:                 r.id,
		HouseholdID:        r.householdID,
		PaidByMemberID:     r.paidByMemberID,
		AmountCents:        r.amountCents,
		Description:        r.description,
		CategoryID:         r.categoryID,
		ExpenseType:        r.expenseType,
		RecurrencePattern:  r.recurrencePattern,
		StartDate:          r.startDate,
		EndDate:            r.endDate,
		NextGenerationDate: r.nextGenerationDate,
		IsActive:           r.isActive,
		CreatedAt:          r.createdAt,
		UpdatedAt:          r.updatedAt,
		DeletedAt:          r.deletedAt,
	}
}

func (r RecurringExpense) ID() ID                               { return r.id }
func (r RecurringExpense) HouseholdID() string                  { return r.householdID }
func (r RecurringExpense) PaidByMemberID() string               { return r.paidByMemberID }
func (r RecurringExpense) AmountCents() int64                   { return r.amountCents }
func (r RecurringExpense) Description() string                  { return r.description }
func (r RecurringExpense) CategoryID() string                   { return r.categoryID }
func (r RecurringExpense) ExpenseType() expense.ExpenseType     { return r.expenseType }
func (r RecurringExpense) RecurrencePattern() RecurrencePattern { return r.recurrencePattern }
func (r RecurringExpense) StartDate() time.Time                 { return r.startDate }
func (r RecurringExpense) EndDate() *time.Time                  { return r.endDate }
func (r RecurringExpense) NextGenerationDate() time.Time        { return r.nextGenerationDate }
func (r RecurringExpense) IsActive() bool                       { return r.isActive }
func (r RecurringExpense) CreatedAt() time.Time                 { return r.createdAt }
func (r RecurringExpense) UpdatedAt() time.Time                 { return r.updatedAt }
func (r RecurringExpense) DeletedAt() *time.Time                { return r.deletedAt }
