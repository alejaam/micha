package recurringexpenseapp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// --- Mocks ---

type mockRecurringExpenseRepo struct {
	outbound.RecurringExpenseRepository
	dueForGeneration []recurringexpense.RecurringExpense
	dueErr           error
	updateCalledWith []recurringexpense.RecurringExpense
	updateErr        error
}

func (m *mockRecurringExpenseRepo) ListDueForGeneration(_ context.Context, _ time.Time) ([]recurringexpense.RecurringExpense, error) {
	return m.dueForGeneration, m.dueErr
}

func (m *mockRecurringExpenseRepo) Update(_ context.Context, re recurringexpense.RecurringExpense) error {
	m.updateCalledWith = append(m.updateCalledWith, re)
	return m.updateErr
}

type mockExpenseRepo struct {
	outbound.ExpenseRepository
	saveErr error
	saved   []expense.Expense
}

func (m *mockExpenseRepo) Save(_ context.Context, e expense.Expense) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, e)
	return nil
}

type mockIDGen struct {
	counter int
}

func (m *mockIDGen) NewID() string {
	m.counter++
	return fmt.Sprintf("mock-id-%d", m.counter)
}

func TestGenerateRecurringExpensesUseCase_Execute(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		recurringExpenses  []recurringexpense.RecurringExpense
		wantGeneratedCount int
		// wantNotSkippedIDs are recurring expense IDs that should have been generated.
		wantNotSkippedIDs []string
	}{
		{
			name: "agnostic fixed expense generates period expense",
			recurringExpenses: []recurringexpense.RecurringExpense{
				mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
					ID:                 "re-agnostic-1",
					HouseholdID:        "hh-1",
					PaidByMemberID:     "",
					IsAgnostic:         true,
					AmountCents:        10000,
					Description:        "Agnostic Service",
					CategoryID:         "cat-1",
					ExpenseType:        expense.ExpenseTypeFixed,
					RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
					StartDate:          startDate,
					NextGenerationDate: now,
					IsActive:           true,
					CreatedAt:          startDate,
					UpdatedAt:          startDate,
				}),
			},
			wantGeneratedCount: 1,
			wantNotSkippedIDs:  []string{"re-agnostic-1"},
		},
		{
			name: "non-agnostic fixed expense still generates",
			recurringExpenses: []recurringexpense.RecurringExpense{
				mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
					ID:                 "re-normal-1",
					HouseholdID:        "hh-1",
					PaidByMemberID:     "member-1",
					IsAgnostic:         false,
					AmountCents:        5000,
					Description:        "Normal Fixed Expense",
					CategoryID:         "cat-1",
					ExpenseType:        expense.ExpenseTypeFixed,
					RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
					StartDate:          startDate,
					NextGenerationDate: now,
					IsActive:           true,
					CreatedAt:          startDate,
					UpdatedAt:          startDate,
				}),
			},
			wantGeneratedCount: 1,
			wantNotSkippedIDs:  []string{"re-normal-1"},
		},
		{
			name: "both agnostic and normal fixed expenses generate",
			recurringExpenses: []recurringexpense.RecurringExpense{
				mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
					ID:                 "re-agnostic-2",
					HouseholdID:        "hh-1",
					PaidByMemberID:     "",
					IsAgnostic:         true,
					AmountCents:        10000,
					Description:        "Agnostic Service",
					CategoryID:         "cat-1",
					ExpenseType:        expense.ExpenseTypeFixed,
					RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
					StartDate:          startDate,
					NextGenerationDate: now,
					IsActive:           true,
					CreatedAt:          startDate,
					UpdatedAt:          startDate,
				}),
				mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
					ID:                 "re-normal-2",
					HouseholdID:        "hh-1",
					PaidByMemberID:     "member-2",
					IsAgnostic:         false,
					AmountCents:        5000,
					Description:        "Normal Fixed Expense",
					CategoryID:         "cat-2",
					ExpenseType:        expense.ExpenseTypeFixed,
					RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
					StartDate:          startDate,
					NextGenerationDate: now,
					IsActive:           true,
					CreatedAt:          startDate,
					UpdatedAt:          startDate,
				}),
			},
			wantGeneratedCount: 2,
			wantNotSkippedIDs:  []string{"re-agnostic-2", "re-normal-2"},
		},
		{
			name: "item not due for generation is skipped",
			recurringExpenses: []recurringexpense.RecurringExpense{
				mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
					ID:                 "re-future-1",
					HouseholdID:        "hh-1",
					PaidByMemberID:     "member-1",
					IsAgnostic:         false,
					AmountCents:        5000,
					Description:        "Future Expense",
					CategoryID:         "cat-1",
					ExpenseType:        expense.ExpenseTypeFixed,
					RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
					StartDate:          now.AddDate(0, 0, 10),
					NextGenerationDate: now.AddDate(0, 0, 10),
					IsActive:           true,
					CreatedAt:          startDate,
					UpdatedAt:          startDate,
				}),
			},
			wantGeneratedCount: 0,
			wantNotSkippedIDs:  nil,
		},
		{
			name: "inactive item is skipped",
			recurringExpenses: []recurringexpense.RecurringExpense{
				mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
					ID:                 "re-inactive-1",
					HouseholdID:        "hh-1",
					PaidByMemberID:     "member-1",
					IsAgnostic:         false,
					AmountCents:        5000,
					Description:        "Inactive Expense",
					CategoryID:         "cat-1",
					ExpenseType:        expense.ExpenseTypeFixed,
					RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
					StartDate:          startDate,
					NextGenerationDate: now,
					IsActive:           false,
					CreatedAt:          startDate,
					UpdatedAt:          startDate,
				}),
			},
			wantGeneratedCount: 0,
			wantNotSkippedIDs:  nil,
		},
		{
			name: "generated expense has fixed type and empty memberID per domain rules",
			recurringExpenses: []recurringexpense.RecurringExpense{
				mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
					ID:                 "re-agnostic-3",
					HouseholdID:        "hh-1",
					PaidByMemberID:     "some-member",
					IsAgnostic:         true,
					AmountCents:        15000,
					Description:        "Agnostic With Member Set",
					CategoryID:         "cat-1",
					ExpenseType:        expense.ExpenseTypeFixed,
					RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
					StartDate:          startDate,
					NextGenerationDate: now,
					IsActive:           true,
					CreatedAt:          startDate,
					UpdatedAt:          startDate,
				}),
			},
			wantGeneratedCount: 1,
			wantNotSkippedIDs:  []string{"re-agnostic-3"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRecurringRepo := &mockRecurringExpenseRepo{
				dueForGeneration: tt.recurringExpenses,
			}
			mockExpRepo := &mockExpenseRepo{}
			uc := NewGenerateRecurringExpensesUseCase(mockRecurringRepo, mockExpRepo, &mockIDGen{})

			out, err := uc.Execute(context.Background(), inbound.GenerateRecurringExpensesCommand{
				AsOfDate: now,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if out.GeneratedCount != tt.wantGeneratedCount {
				t.Errorf("generated count = %d; want %d", out.GeneratedCount, tt.wantGeneratedCount)
			}

			if len(out.ExpenseIDs) != tt.wantGeneratedCount {
				t.Errorf("expense IDs count = %d; want %d", len(out.ExpenseIDs), tt.wantGeneratedCount)
			}

			if len(mockExpRepo.saved) != tt.wantGeneratedCount {
				t.Errorf("saved count = %d; want %d", len(mockExpRepo.saved), tt.wantGeneratedCount)
			}

			// Verify that generated expenses have the expected properties.
			// Domain rule: for ExpenseTypeFixed, paid_by_member_id is always "".
			for i, saved := range mockExpRepo.saved {
				if saved.ExpenseType() != expense.ExpenseTypeFixed {
					t.Errorf("saved[%d] expense_type = %q; want %q", i, saved.ExpenseType(), expense.ExpenseTypeFixed)
				}
				if saved.PaidByMemberID() != "" {
					t.Errorf("saved[%d] paid_by_member_id = %q; want empty string (domain rule for fixed)", i, saved.PaidByMemberID())
				}
				// Verify it's shared (as set by the use case)
				if !saved.IsShared() {
					t.Errorf("saved[%d] is_shared = false; want true", i)
				}
				// Verify description suffix
				if len(tt.wantNotSkippedIDs) > 0 && i < len(tt.recurringExpenses) {
					reDesc := tt.recurringExpenses[i].Description()
					wantDesc := fmt.Sprintf("%s (auto)", reDesc)
					if saved.Description() != wantDesc {
						t.Errorf("saved[%d] description = %q; want %q", i, saved.Description(), wantDesc)
					}
				}
			}
		})
	}
}

func TestGenerateRecurringExpensesUseCase_WithHouseholdFilter(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	items := []recurringexpense.RecurringExpense{
		mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
			ID:                 "re-hh1",
			HouseholdID:        "hh-1",
			PaidByMemberID:     "member-1",
			IsAgnostic:         false,
			AmountCents:        5000,
			Description:        "HH1 Expense",
			CategoryID:         "cat-1",
			ExpenseType:        expense.ExpenseTypeFixed,
			RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
			StartDate:          startDate,
			NextGenerationDate: now,
			IsActive:           true,
			CreatedAt:          startDate,
			UpdatedAt:          startDate,
		}),
		mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
			ID:                 "re-hh2",
			HouseholdID:        "hh-2",
			PaidByMemberID:     "member-2",
			IsAgnostic:         false,
			AmountCents:        3000,
			Description:        "HH2 Expense",
			CategoryID:         "cat-1",
			ExpenseType:        expense.ExpenseTypeFixed,
			RecurrencePattern:  recurringexpense.RecurrencePatternMonthly,
			StartDate:          startDate,
			NextGenerationDate: now,
			IsActive:           true,
			CreatedAt:          startDate,
			UpdatedAt:          startDate,
		}),
	}

	mockRecurringRepo := &mockRecurringExpenseRepo{
		dueForGeneration: items,
	}
	mockExpRepo := &mockExpenseRepo{}
	uc := NewGenerateRecurringExpensesUseCase(mockRecurringRepo, mockExpRepo, &mockIDGen{})

	out, err := uc.Execute(context.Background(), inbound.GenerateRecurringExpensesCommand{
		HouseholdID: "hh-1",
		AsOfDate:    now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.GeneratedCount != 1 {
		t.Errorf("generated count = %d; want 1", out.GeneratedCount)
	}

	if len(mockExpRepo.saved) != 1 {
		t.Fatalf("saved count = %d; want 1", len(mockExpRepo.saved))
	}

	if mockExpRepo.saved[0].HouseholdID() != "hh-1" {
		t.Errorf("saved expense household = %q; want %q", mockExpRepo.saved[0].HouseholdID(), "hh-1")
	}
}

func mustNewRecurringExpense(t *testing.T, attrs recurringexpense.RecurringExpenseAttributes) recurringexpense.RecurringExpense {
	t.Helper()
	re, err := recurringexpense.NewFromAttributes(attrs)
	if err != nil {
		t.Fatalf("failed to create recurring expense: %v", err)
	}
	return re
}
