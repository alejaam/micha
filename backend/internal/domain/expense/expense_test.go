package expense_test

import (
	"errors"
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
)

var baseAttrs = expense.ExpenseAttributes{
	ID:             expense.ID("id-1"),
	HouseholdID:    "hh-1",
	PaidByMemberID: "m-1",
	AmountCents:    1000,
	Description:    "Taxi",
	IsShared:       true,
	Currency:       "MXN",
	PaymentMethod:  expense.PaymentMethodCard,
	ExpenseType:    expense.ExpenseTypeVariable,
	CreatedAt:      time.Now(),
}

func TestNew_ValidExpense(t *testing.T) {
	t.Parallel()
	e, err := expense.NewFromAttributes(baseAttrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.AmountCents() != baseAttrs.AmountCents {
		t.Errorf("AmountCents = %d; want %d", e.AmountCents(), baseAttrs.AmountCents)
	}
}

func TestNew_InvalidMoney(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		amountCents int64
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			attrs := baseAttrs
			attrs.AmountCents = tc.amountCents
			_, err := expense.NewFromAttributes(attrs)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, shared.ErrInvalidMoney) {
				t.Errorf("want ErrInvalidMoney, got %v", err)
			}
		})
	}
}

func TestPatch(t *testing.T) {
	t.Parallel()
	t.Run("patch amount", func(t *testing.T) {
		t.Parallel()
		e := mustNew(t, baseAttrs)
		newAmount := int64(2000)
		if err := e.Patch(nil, &newAmount); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.AmountCents() != 2000 {
			t.Errorf("AmountCents = %d; want 2000", e.AmountCents())
		}
	})
	t.Run("patch description", func(t *testing.T) {
		t.Parallel()
		e := mustNew(t, baseAttrs)
		desc := "Updated"
		if err := e.Patch(&desc, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Description() != "Updated" {
			t.Errorf("Description = %q; want %q", e.Description(), "Updated")
		}
	})
	t.Run("patch with zero amount rejects", func(t *testing.T) {
		t.Parallel()
		e := mustNew(t, baseAttrs)
		bad := int64(0)
		err := e.Patch(nil, &bad)
		if !errors.Is(err, shared.ErrInvalidMoney) {
			t.Errorf("want ErrInvalidMoney, got %v", err)
		}
	})
	t.Run("updatedAt advances", func(t *testing.T) {
		t.Parallel()
		e := mustNew(t, baseAttrs)
		before := e.UpdatedAt()
		time.Sleep(time.Millisecond)
		amount := int64(500)
		_ = e.Patch(nil, &amount)
		if !e.UpdatedAt().After(before) {
			t.Errorf("UpdatedAt did not advance: before=%v after=%v", before, e.UpdatedAt())
		}
	})
}

func TestNew_ExpenseTypeDefaultsToVariable(t *testing.T) {
	t.Parallel()
	attrs := baseAttrs
	attrs.ExpenseType = ""

	e, err := expense.NewFromAttributes(attrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.ExpenseType() != expense.ExpenseTypeVariable {
		t.Errorf("ExpenseType = %q; want %q", e.ExpenseType(), expense.ExpenseTypeVariable)
	}
}

func TestNew_InvalidExpenseType(t *testing.T) {
	t.Parallel()
	attrs := baseAttrs
	attrs.ExpenseType = expense.ExpenseType("unknown")

	_, err := expense.NewFromAttributes(attrs)
	if !errors.Is(err, expense.ErrInvalidExpenseType) {
		t.Errorf("want ErrInvalidExpenseType, got %v", err)
	}
}

func TestSoftDelete(t *testing.T) {
	t.Parallel()
	t.Run("sets deletedAt", func(t *testing.T) {
		t.Parallel()
		e := mustNew(t, baseAttrs)
		if err := e.SoftDelete(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.DeletedAt() == nil {
			t.Error("DeletedAt should not be nil after SoftDelete")
		}
	})
	t.Run("double delete returns ErrAlreadyDeleted", func(t *testing.T) {
		t.Parallel()
		e := mustNew(t, baseAttrs)
		_ = e.SoftDelete()
		err := e.SoftDelete()
		if !errors.Is(err, shared.ErrAlreadyDeleted) {
			t.Errorf("want ErrAlreadyDeleted, got %v", err)
		}
	})
}

func TestAttributes_RoundTrip(t *testing.T) {
	t.Parallel()
	e := mustNew(t, baseAttrs)
	attrs := e.Attributes()
	e2, err := expense.NewFromAttributes(attrs)
	if err != nil {
		t.Fatalf("rehydration error: %v", err)
	}
	if e2.ID() != e.ID() {
		t.Errorf("ID mismatch: %v != %v", e2.ID(), e.ID())
	}
}

func mustNew(t *testing.T, attrs expense.ExpenseAttributes) *expense.Expense {
	t.Helper()
	e, err := expense.NewFromAttributes(attrs)
	if err != nil {
		t.Fatalf("mustNew: %v", err)
	}
	return &e
}
