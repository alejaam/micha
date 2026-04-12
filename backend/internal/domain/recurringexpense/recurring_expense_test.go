package recurringexpense_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/domain/shared"
)

func TestNew(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		id                recurringexpense.ID
		householdID       string
		paidByMemberID    string
		amountCents       int64
		description       string
		categoryID        string
		expenseType       expense.ExpenseType
		recurrencePattern recurringexpense.RecurrencePattern
		startDate         time.Time
		createdAt         time.Time
		wantErr           error
	}{
		{
			name:              "valid monthly recurring expense",
			id:                "rec-1",
			householdID:       "hh-1",
			paidByMemberID:    "member-1",
			amountCents:       50000,
			description:       "Monthly rent",
			categoryID:        "cat-rent",
			expenseType:       expense.ExpenseTypeFixed,
			recurrencePattern: recurringexpense.RecurrencePatternMonthly,
			startDate:         startDate,
			createdAt:         now,
			wantErr:           nil,
		},
		{
			name:              "valid biweekly recurring expense",
			id:                "rec-2",
			householdID:       "hh-1",
			paidByMemberID:    "member-1",
			amountCents:       15000,
			description:       "Biweekly groceries",
			categoryID:        "cat-food",
			expenseType:       expense.ExpenseTypeVariable,
			recurrencePattern: recurringexpense.RecurrencePatternBiweekly,
			startDate:         startDate,
			createdAt:         now,
			wantErr:           nil,
		},
		{
			name:              "valid weekly recurring expense",
			id:                "rec-3",
			householdID:       "hh-1",
			paidByMemberID:    "member-1",
			amountCents:       5000,
			description:       "Weekly transportation",
			categoryID:        "cat-auto",
			expenseType:       expense.ExpenseTypeVariable,
			recurrencePattern: recurringexpense.RecurrencePatternWeekly,
			startDate:         startDate,
			createdAt:         now,
			wantErr:           nil,
		},
		{
			name:              "invalid amount",
			id:                "rec-4",
			householdID:       "hh-1",
			paidByMemberID:    "member-1",
			amountCents:       0,
			description:       "Invalid",
			categoryID:        "cat-other",
			expenseType:       expense.ExpenseTypeFixed,
			recurrencePattern: recurringexpense.RecurrencePatternMonthly,
			startDate:         startDate,
			createdAt:         now,
			wantErr:           shared.ErrInvalidMoney,
		},
		{
			name:              "invalid household id",
			id:                "rec-5",
			householdID:       "",
			paidByMemberID:    "member-1",
			amountCents:       10000,
			description:       "Invalid",
			categoryID:        "cat-other",
			expenseType:       expense.ExpenseTypeFixed,
			recurrencePattern: recurringexpense.RecurrencePatternMonthly,
			startDate:         startDate,
			createdAt:         now,
			wantErr:           recurringexpense.ErrInvalidHouseholdID,
		},
		{
			name:              "invalid paid by member id",
			id:                "rec-6",
			householdID:       "hh-1",
			paidByMemberID:    "",
			amountCents:       10000,
			description:       "Invalid",
			categoryID:        "cat-other",
			expenseType:       expense.ExpenseTypeFixed,
			recurrencePattern: recurringexpense.RecurrencePatternMonthly,
			startDate:         startDate,
			createdAt:         now,
			wantErr:           recurringexpense.ErrInvalidPaidByMemberID,
		},
		{
			name:              "invalid recurrence pattern",
			id:                "rec-7",
			householdID:       "hh-1",
			paidByMemberID:    "member-1",
			amountCents:       10000,
			description:       "Invalid",
			categoryID:        "cat-other",
			expenseType:       expense.ExpenseTypeFixed,
			recurrencePattern: "daily",
			startDate:         startDate,
			createdAt:         now,
			wantErr:           recurringexpense.ErrInvalidRecurrencePattern,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := recurringexpense.New(
				tt.id,
				tt.householdID,
				tt.paidByMemberID,
				tt.amountCents,
				tt.description,
				tt.categoryID,
				tt.expenseType,
				tt.recurrencePattern,
				tt.startDate,
				tt.createdAt,
			)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.id, got.ID())
			assert.Equal(t, tt.householdID, got.HouseholdID())
			assert.Equal(t, tt.paidByMemberID, got.PaidByMemberID())
			assert.Equal(t, tt.amountCents, got.AmountCents())
			assert.Equal(t, tt.description, got.Description())
			assert.Equal(t, tt.categoryID, got.CategoryID())
			assert.Equal(t, tt.expenseType, got.ExpenseType())
			assert.Equal(t, tt.recurrencePattern, got.RecurrencePattern())
			assert.Equal(t, tt.startDate, got.StartDate())
			assert.Nil(t, got.EndDate())
			assert.Equal(t, tt.startDate, got.NextGenerationDate())
			assert.True(t, got.IsActive())
			assert.Equal(t, tt.createdAt, got.CreatedAt())
			assert.Nil(t, got.DeletedAt())
		})
	}
}

func TestNewFromAttributes_AllowsAgnosticWithoutPaidByMember(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	re, err := recurringexpense.NewFromAttributes(recurringexpense.RecurringExpenseAttributes{
		ID:                 "rec-agnostic-ok",
		HouseholdID:        "hh-1",
		PaidByMemberID:     "",
		IsAgnostic:         true,
		AmountCents:        25000,
		Description:        "Internet",
		CategoryID:         "cat-other",
		ExpenseType:        expense.ExpenseTypeFixed,
		RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
		StartDate:          startDate,
		NextGenerationDate: startDate,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	require.NoError(t, err)
	assert.True(t, re.IsAgnostic())
	assert.Equal(t, "", re.PaidByMemberID())
}

func TestNewFromAttributes_AgnosticMustBeFixed(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	_, err := recurringexpense.NewFromAttributes(recurringexpense.RecurringExpenseAttributes{
		ID:                 "rec-agnostic-invalid",
		HouseholdID:        "hh-1",
		PaidByMemberID:     "",
		IsAgnostic:         true,
		AmountCents:        25000,
		Description:        "Invalid",
		CategoryID:         "cat-other",
		ExpenseType:        expense.ExpenseTypeVariable,
		RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
		StartDate:          startDate,
		NextGenerationDate: startDate,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, recurringexpense.ErrAgnosticRequiresFixedType)
}

func TestNewFromAttributes(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		attrs   recurringexpense.RecurringExpenseAttributes
		wantErr error
	}{
		{
			name: "valid with end date",
			attrs: recurringexpense.RecurringExpenseAttributes{
				ID:                 "rec-1",
				HouseholdID:        "hh-1",
				PaidByMemberID:     "member-1",
				AmountCents:        50000,
				Description:        "Monthly rent",
				CategoryID:         "cat-rent",
				ExpenseType:        expense.ExpenseTypeFixed,
				RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
				StartDate:          startDate,
				EndDate:            &endDate,
				NextGenerationDate: startDate,
				IsActive:           true,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
			wantErr: nil,
		},
		{
			name: "invalid end date before start date",
			attrs: recurringexpense.RecurringExpenseAttributes{
				ID:                 "rec-2",
				HouseholdID:        "hh-1",
				PaidByMemberID:     "member-1",
				AmountCents:        50000,
				Description:        "Invalid",
				CategoryID:         "cat-rent",
				ExpenseType:        expense.ExpenseTypeFixed,
				RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
				StartDate:          endDate,
				EndDate:            &startDate,
				NextGenerationDate: endDate,
				IsActive:           true,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
			wantErr: recurringexpense.ErrInvalidDateRange,
		},
		{
			name: "invalid next generation date before start date",
			attrs: recurringexpense.RecurringExpenseAttributes{
				ID:                 "rec-3",
				HouseholdID:        "hh-1",
				PaidByMemberID:     "member-1",
				AmountCents:        50000,
				Description:        "Invalid",
				CategoryID:         "cat-rent",
				ExpenseType:        expense.ExpenseTypeFixed,
				RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
				StartDate:          startDate,
				EndDate:            nil,
				NextGenerationDate: startDate.AddDate(0, -1, 0),
				IsActive:           true,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
			wantErr: recurringexpense.ErrInvalidNextGenerationDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := recurringexpense.NewFromAttributes(tt.attrs)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.attrs.ID, got.ID())
			assert.Equal(t, tt.attrs.EndDate, got.EndDate())
		})
	}
}

func TestPatch(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	re, err := recurringexpense.New(
		"rec-1",
		"hh-1",
		"member-1",
		50000,
		"Monthly rent",
		"cat-rent",
		expense.ExpenseTypeFixed,
		recurringexpense.RecurrencePatternMonthly,
		startDate,
		now,
	)
	require.NoError(t, err)

	newDesc := "Updated rent"
	newAmount := int64(60000)
	newCategory := "cat-housing"
	newActive := false

	err = re.Patch(&newDesc, &newAmount, &newCategory, &newActive)
	require.NoError(t, err)

	assert.Equal(t, newDesc, re.Description())
	assert.Equal(t, newAmount, re.AmountCents())
	assert.Equal(t, newCategory, re.CategoryID())
	assert.False(t, re.IsActive())

	// Test invalid amount
	invalidAmount := int64(-100)
	err = re.Patch(nil, &invalidAmount, nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrInvalidMoney)
}

func TestAdvanceNextGenerationDate(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		recurrencePattern recurringexpense.RecurrencePattern
		expectedDate      time.Time
	}{
		{
			name:              "monthly advances by 1 month",
			recurrencePattern: recurringexpense.RecurrencePatternMonthly,
			expectedDate:      time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:              "biweekly advances by 14 days",
			recurrencePattern: recurringexpense.RecurrencePatternBiweekly,
			expectedDate:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:              "weekly advances by 7 days",
			recurrencePattern: recurringexpense.RecurrencePatternWeekly,
			expectedDate:      time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := recurringexpense.New(
				"rec-1",
				"hh-1",
				"member-1",
				50000,
				"Test",
				"cat-test",
				expense.ExpenseTypeFixed,
				tt.recurrencePattern,
				startDate,
				now,
			)
			require.NoError(t, err)

			re.AdvanceNextGenerationDate()
			assert.Equal(t, tt.expectedDate, re.NextGenerationDate())
		})
	}
}

func TestShouldGenerateExpense(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		attrs    recurringexpense.RecurringExpenseAttributes
		checkAt  time.Time
		expected bool
	}{
		{
			name: "active and due",
			attrs: recurringexpense.RecurringExpenseAttributes{
				ID:                 "rec-1",
				HouseholdID:        "hh-1",
				PaidByMemberID:     "member-1",
				AmountCents:        50000,
				Description:        "Test",
				CategoryID:         "cat-test",
				ExpenseType:        expense.ExpenseTypeFixed,
				RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
				StartDate:          startDate,
				EndDate:            nil,
				NextGenerationDate: startDate,
				IsActive:           true,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
			checkAt:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name: "not yet due",
			attrs: recurringexpense.RecurringExpenseAttributes{
				ID:                 "rec-2",
				HouseholdID:        "hh-1",
				PaidByMemberID:     "member-1",
				AmountCents:        50000,
				Description:        "Test",
				CategoryID:         "cat-test",
				ExpenseType:        expense.ExpenseTypeFixed,
				RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
				StartDate:          startDate,
				EndDate:            nil,
				NextGenerationDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				IsActive:           true,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
			checkAt:  time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name: "inactive",
			attrs: recurringexpense.RecurringExpenseAttributes{
				ID:                 "rec-3",
				HouseholdID:        "hh-1",
				PaidByMemberID:     "member-1",
				AmountCents:        50000,
				Description:        "Test",
				CategoryID:         "cat-test",
				ExpenseType:        expense.ExpenseTypeFixed,
				RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
				StartDate:          startDate,
				EndDate:            nil,
				NextGenerationDate: startDate,
				IsActive:           false,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
			checkAt:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name: "after end date",
			attrs: recurringexpense.RecurringExpenseAttributes{
				ID:                 "rec-4",
				HouseholdID:        "hh-1",
				PaidByMemberID:     "member-1",
				AmountCents:        50000,
				Description:        "Test",
				CategoryID:         "cat-test",
				ExpenseType:        expense.ExpenseTypeFixed,
				RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
				StartDate:          startDate,
				EndDate:            &endDate,
				NextGenerationDate: startDate,
				IsActive:           true,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
			checkAt:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re, err := recurringexpense.NewFromAttributes(tt.attrs)
			require.NoError(t, err)

			result := re.ShouldGenerateExpense(tt.checkAt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSoftDelete(t *testing.T) {
	now := time.Now()
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	re, err := recurringexpense.New(
		"rec-1",
		"hh-1",
		"member-1",
		50000,
		"Test",
		"cat-test",
		expense.ExpenseTypeFixed,
		recurringexpense.RecurrencePatternMonthly,
		startDate,
		now,
	)
	require.NoError(t, err)

	assert.Nil(t, re.DeletedAt())

	err = re.SoftDelete()
	require.NoError(t, err)
	assert.NotNil(t, re.DeletedAt())

	// Second delete should fail
	err = re.SoftDelete()
	require.Error(t, err)
	assert.ErrorIs(t, err, shared.ErrAlreadyDeleted)
}
