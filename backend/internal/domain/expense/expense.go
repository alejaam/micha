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
	ErrInvalidExpenseType    = errors.New("invalid expense type")
	ErrInvalidCategory       = errors.New("invalid expense category")
)

// PaymentMethod defines how an expense was paid.
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodVoucher  PaymentMethod = "voucher"
)

// ExpenseType defines the planning category of an expense.
type ExpenseType string

const (
	ExpenseTypeFixed    ExpenseType = "fixed"
	ExpenseTypeVariable ExpenseType = "variable"
	ExpenseTypeMSI      ExpenseType = "msi"
)

// Category is the semantic grouping of an expense (mirrors the Excel panels).
type Category string

const (
	CategoryRent      Category = "rent"
	CategoryAuto      Category = "auto"
	CategoryStreaming Category = "streaming"
	CategoryFood      Category = "food"
	CategoryPersonal  Category = "personal"
	CategorySavings   Category = "savings"
	CategoryOther     Category = "other"
)

// ID is the unique identifier type for an expense.
type ID string

// ExpenseAttributes is the flat DTO used for construction and rehydration.
type ExpenseAttributes struct {
	ID                ID
	PaidByMemberID    string // The member who paid this expense
	PeriodID          string // The period this expense belongs to
	CategoryID        string // The category this expense belongs to
	CardID            string // Optional linked card id when payment_method=card
	HouseholdID       string
	AmountCents       int64
	Description       string
	IsShared          bool
	Currency          string
	PaymentMethod     PaymentMethod
	ExpenseType       ExpenseType
	CardName          string
	TotalInstallments int // Only for ExpenseTypeMSI; number of installments (e.g., 3, 6, 12)
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// Expense is the aggregate root for an expense record.
type Expense struct {
	id                ID
	memberID          string
	periodID          string
	categoryID        string
	cardID            string
	householdID       string
	amountCents       int64
	description       string
	isShared          bool
	currency          string
	paymentMethod     PaymentMethod
	expenseType       ExpenseType
	cardName          string
	totalInstallments int
	createdAt         time.Time
	updatedAt         time.Time
	deletedAt         *time.Time
}

// New constructs an Expense from individual fields.
func New(id ID, householdID string, amountCents int64, description string, createdAt time.Time) (Expense, error) {
	return NewFromAttributes(ExpenseAttributes{
		ID:             id,
		HouseholdID:    householdID,
		PaidByMemberID: "unassigned",
		PeriodID:       "",
		CategoryID:     "",
		AmountCents:    amountCents,
		Description:    description,
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  PaymentMethodCash,
		ExpenseType:    ExpenseTypeVariable,
		CardName:       "",
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

	memberID := strings.TrimSpace(attrs.PaidByMemberID)
	if memberID == "" {
		return Expense{}, ErrInvalidPaidByMemberID
	}

	// PeriodID and CategoryID are optional during construction
	// They will be validated by the application layer

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

	expenseType := attrs.ExpenseType
	if expenseType == "" {
		expenseType = ExpenseTypeVariable
	}
	if expenseType != ExpenseTypeFixed && expenseType != ExpenseTypeVariable && expenseType != ExpenseTypeMSI {
		return Expense{}, ErrInvalidExpenseType
	}

	// Validate TotalInstallments for MSI expenses
	totalInstallments := attrs.TotalInstallments
	if expenseType == ExpenseTypeMSI && totalInstallments <= 0 {
		return Expense{}, errors.New("MSI expense requires total_installments > 0")
	}

	updatedAt := attrs.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = attrs.CreatedAt
	}

	return Expense{
		id:                attrs.ID,
		memberID:          memberID,
		periodID:          strings.TrimSpace(attrs.PeriodID),
		categoryID:        strings.TrimSpace(attrs.CategoryID),
		cardID:            strings.TrimSpace(attrs.CardID),
		householdID:       attrs.HouseholdID,
		amountCents:       attrs.AmountCents,
		description:       attrs.Description,
		isShared:          attrs.IsShared,
		currency:          currency,
		paymentMethod:     paymentMethod,
		expenseType:       expenseType,
		cardName:          strings.TrimSpace(attrs.CardName),
		totalInstallments: totalInstallments,
		createdAt:         attrs.CreatedAt,
		updatedAt:         updatedAt,
		deletedAt:         attrs.DeletedAt,
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
		ID:                e.id,
		PaidByMemberID:    e.memberID,
		PeriodID:          e.periodID,
		CategoryID:        e.categoryID,
		CardID:            e.cardID,
		HouseholdID:       e.householdID,
		AmountCents:       e.amountCents,
		Description:       e.description,
		IsShared:          e.isShared,
		Currency:          e.currency,
		PaymentMethod:     e.paymentMethod,
		ExpenseType:       e.expenseType,
		CardName:          e.cardName,
		TotalInstallments: e.totalInstallments,
		CreatedAt:         e.createdAt,
		UpdatedAt:         e.updatedAt,
		DeletedAt:         e.deletedAt,
	}
}

func (e Expense) ID() ID                       { return e.id }
func (e Expense) PaidByMemberID() string       { return e.memberID }
func (e Expense) PeriodID() string             { return e.periodID }
func (e Expense) CategoryID() string           { return e.categoryID }
func (e Expense) CardID() string               { return e.cardID }
func (e Expense) HouseholdID() string          { return e.householdID }
func (e Expense) AmountCents() int64           { return e.amountCents }
func (e Expense) Description() string          { return e.description }
func (e Expense) IsShared() bool               { return e.isShared }
func (e Expense) Currency() string             { return e.currency }
func (e Expense) PaymentMethod() PaymentMethod { return e.paymentMethod }
func (e Expense) ExpenseType() ExpenseType     { return e.expenseType }
func (e Expense) CardName() string             { return e.cardName }
func (e Expense) TotalInstallments() int       { return e.totalInstallments }
func (e Expense) CreatedAt() time.Time         { return e.createdAt }
func (e Expense) UpdatedAt() time.Time         { return e.updatedAt }
func (e Expense) DeletedAt() *time.Time        { return e.deletedAt }

// Legacy method for backward compatibility
func (e Expense) MemberID() string { return e.memberID }
