package installment

import (
	"testing"
	"time"

	"micha/backend/internal/domain/shared"

	"github.com/stretchr/testify/assert"
)

// RED: Test for new fields in InstallmentAttributes
func TestNewFromAttributes_WithNewFields(t *testing.T) {
	now := time.Now()
	attrs := InstallmentAttributes{
		ID:                     "inst-1",
		ExpenseID:              "exp-1",
		PaidByMemberID:         "member-1", // NEW field
		StartDate:              now,        // NEW field
		InstallmentAmountCents: 334,        // NEW field
		TotalAmountCents:       1000,
		TotalInstallments:      3,
		CurrentInstallment:     1,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	installment, err := NewFromAttributes(attrs)

	assert.NoError(t, err)
	assert.Equal(t, ID("inst-1"), installment.ID())
	assert.Equal(t, "exp-1", installment.ExpenseID())
	assert.Equal(t, "member-1", installment.PaidByMemberID())         // NEW getter
	assert.Equal(t, now, installment.StartDate())                     // Already exists
	assert.Equal(t, int64(334), installment.InstallmentAmountCents()) // NEW getter
	assert.Equal(t, int64(1000), installment.TotalAmountCents())
}

// TRIANGULATE: Test validation for empty PaidByMemberID
func TestNewFromAttributes_EmptyPaidByMemberID(t *testing.T) {
	now := time.Now()
	attrs := InstallmentAttributes{
		ID:                     "inst-1",
		ExpenseID:              "exp-1",
		PaidByMemberID:         "", // Invalid
		StartDate:              now,
		InstallmentAmountCents: 334,
		TotalAmountCents:       1000,
		TotalInstallments:      3,
		CurrentInstallment:     1,
		CreatedAt:              now,
	}

	_, err := NewFromAttributes(attrs)

	assert.ErrorIs(t, err, shared.ErrInvalidID)
}

// TRIANGULATE: Test validation for zero InstallmentAmountCents
func TestNewFromAttributes_ZeroInstallmentAmount(t *testing.T) {
	now := time.Now()
	attrs := InstallmentAttributes{
		ID:                     "inst-1",
		ExpenseID:              "exp-1",
		PaidByMemberID:         "member-1",
		StartDate:              now,
		InstallmentAmountCents: 0, // Invalid
		TotalAmountCents:       1000,
		TotalInstallments:      3,
		CurrentInstallment:     1,
		CreatedAt:              now,
	}

	_, err := NewFromAttributes(attrs)

	assert.ErrorIs(t, err, shared.ErrInvalidMoney)
}

// TRIANGULATE: Test with different installment amounts (remainder distribution)
func TestNewFromAttributes_DifferentAmounts(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		amountCents    int64
		expectedAmount int64
	}{
		{"first installment with remainder", 334, 334},
		{"subsequent installment no remainder", 333, 333},
		{"large amount", 40000, 40000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs := InstallmentAttributes{
				ID:                     "inst-1",
				ExpenseID:              "exp-1",
				PaidByMemberID:         "member-1",
				StartDate:              now,
				InstallmentAmountCents: tt.amountCents,
				TotalAmountCents:       1000,
				TotalInstallments:      3,
				CurrentInstallment:     1,
				CreatedAt:              now,
			}

			installment, err := NewFromAttributes(attrs)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAmount, installment.InstallmentAmountCents())
		})
	}
}
