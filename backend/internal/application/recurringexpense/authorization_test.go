package recurringexpenseapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type authorizationRecurringExpenseRepo struct {
	outbound.RecurringExpenseRepository
	re                     recurringexpense.RecurringExpense
	findCalls              int
	listCalls              int
	listDueForGenerationCalls int
	saveCalls              int
	updateCalls            int
}

func (r *authorizationRecurringExpenseRepo) FindByID(_ context.Context, _ string) (recurringexpense.RecurringExpense, error) {
	r.findCalls++
	return r.re, nil
}

func (r *authorizationRecurringExpenseRepo) List(_ context.Context, _ string, _, _ int) ([]recurringexpense.RecurringExpense, error) {
	r.listCalls++
	return nil, nil
}

func (r *authorizationRecurringExpenseRepo) ListDueForGeneration(_ context.Context, _ time.Time) ([]recurringexpense.RecurringExpense, error) {
	r.listDueForGenerationCalls++
	return nil, nil
}

func (r *authorizationRecurringExpenseRepo) Save(_ context.Context, _ recurringexpense.RecurringExpense) error {
	r.saveCalls++
	return nil
}

func (r *authorizationRecurringExpenseRepo) Update(_ context.Context, _ recurringexpense.RecurringExpense) error {
	r.updateCalls++
	return nil
}

type authorizationMemberRepo struct {
	outbound.MemberRepository
	err         error
	calls       int
	householdID string
	userID      string
}

func (r *authorizationMemberRepo) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	r.calls++
	r.householdID = householdID
	r.userID = userID
	return member.Member{}, r.err
}

type authorizationHouseholdRepo struct {
	outbound.HouseholdRepository
}

func (r authorizationHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	return household.New(household.ID(id), "Test household", "owner-1", household.SettlementModeEqual, "MXN", time.Now())
}

func TestRecurringExpenseUseCases_RejectNonMembersBeforeTargetRepositoryAccess(t *testing.T) {
	t.Parallel()

	re := mustNewRecurringExpense(t, recurringexpense.RecurringExpenseAttributes{
		ID:                 "re-1",
		HouseholdID:        "household-1",
		PaidByMemberID:     "member-1",
		AmountCents:        1000,
		Description:        "Rent",
		CategoryID:         "category-1",
		ExpenseType:        "fixed",
		RecurrencePattern:  "monthly",
		StartDate:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		NextGenerationDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		IsActive:           true,
		CreatedAt:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	tests := []struct {
		name       string
		execute    func(*authorizationRecurringExpenseRepo, *authorizationMemberRepo) error
		assertRepo func(*testing.T, *authorizationRecurringExpenseRepo)
	}{
		{
			name: "create",
			execute: func(repo *authorizationRecurringExpenseRepo, members *authorizationMemberRepo) error {
				_, err := NewCreateRecurringExpenseUseCase(repo, authorizationHouseholdRepo{}, members, nil, &mockIDGen{}).Execute(context.Background(), inbound.CreateRecurringExpenseInput{
					CurrentUserID:     "user-1",
					HouseholdID:       "household-1",
					PaidByMemberID:    "member-1",
					AmountCents:       1000,
					Description:       "Rent",
					CategoryID:        "category-1",
					ExpenseType:       "fixed",
					RecurrencePattern: "monthly",
					StartDate:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				})
				return err
			},
			assertRepo: func(t *testing.T, repo *authorizationRecurringExpenseRepo) {
				t.Helper()
				if repo.saveCalls != 0 {
					t.Fatalf("Save calls = %d; want 0", repo.saveCalls)
				}
			},
		},
		{
			name: "get resolves actual household before authorization",
			execute: func(repo *authorizationRecurringExpenseRepo, members *authorizationMemberRepo) error {
				_, err := NewGetRecurringExpenseUseCase(repo, members).Execute(context.Background(), inbound.GetRecurringExpenseInput{ID: "re-1", CurrentUserID: "user-1"})
				return err
			},
			assertRepo: func(t *testing.T, repo *authorizationRecurringExpenseRepo) {
				t.Helper()
				if repo.findCalls != 1 {
					t.Fatalf("FindByID calls = %d; want 1", repo.findCalls)
				}
			},
		},
		{
			name: "list authorizes requested household before listing",
			execute: func(repo *authorizationRecurringExpenseRepo, members *authorizationMemberRepo) error {
				_, err := NewListRecurringExpensesUseCase(repo, members).Execute(context.Background(), inbound.ListRecurringExpensesQuery{HouseholdID: "household-1", CurrentUserID: "user-1"})
				return err
			},
			assertRepo: func(t *testing.T, repo *authorizationRecurringExpenseRepo) {
				t.Helper()
				if repo.listCalls != 0 {
					t.Fatalf("List calls = %d; want 0", repo.listCalls)
				}
			},
		},
		{
			name: "update resolves actual household before mutation",
			execute: func(repo *authorizationRecurringExpenseRepo, members *authorizationMemberRepo) error {
				_, err := NewUpdateRecurringExpenseUseCase(repo, members, nil).Execute(context.Background(), inbound.UpdateRecurringExpenseCommand{ID: "re-1", CurrentUserID: "user-1"})
				return err
			},
			assertRepo: func(t *testing.T, repo *authorizationRecurringExpenseRepo) {
				t.Helper()
				if repo.updateCalls != 0 {
					t.Fatalf("Update calls = %d; want 0", repo.updateCalls)
				}
			},
		},
		{
			name: "delete resolves actual household before mutation",
			execute: func(repo *authorizationRecurringExpenseRepo, members *authorizationMemberRepo) error {
				return NewDeleteRecurringExpenseUseCase(repo, members).Execute(context.Background(), inbound.DeleteRecurringExpenseInput{ID: "re-1", CurrentUserID: "user-1"})
			},
			assertRepo: func(t *testing.T, repo *authorizationRecurringExpenseRepo) {
				t.Helper()
				if repo.updateCalls != 0 {
					t.Fatalf("Update calls = %d; want 0", repo.updateCalls)
				}
			},
		},
		{
			name: "manual generation authorizes requested household before listing due expenses",
			execute: func(repo *authorizationRecurringExpenseRepo, members *authorizationMemberRepo) error {
				_, err := NewGenerateRecurringExpensesUseCase(repo, members, &mockExpenseRepo{}, &mockIDGen{}).Execute(context.Background(), inbound.GenerateRecurringExpensesCommand{HouseholdID: "household-1", CurrentUserID: "user-1"})
				return err
			},
			assertRepo: func(t *testing.T, repo *authorizationRecurringExpenseRepo) {
				t.Helper()
				if repo.listDueForGenerationCalls != 0 {
					t.Fatalf("ListDueForGeneration calls = %d; want 0", repo.listDueForGenerationCalls)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &authorizationRecurringExpenseRepo{re: re}
			members := &authorizationMemberRepo{err: shared.ErrNotFound}

			err := tt.execute(repo, members)
			if !errors.Is(err, shared.ErrForbidden) {
				t.Fatalf("error = %v; want ErrForbidden", err)
			}
			if members.calls != 1 {
				t.Fatalf("FindByUserID calls = %d; want 1", members.calls)
			}
			if members.householdID != "household-1" || members.userID != "user-1" {
				t.Fatalf("FindByUserID(%q, %q); want (household-1, user-1)", members.householdID, members.userID)
			}
			tt.assertRepo(t, repo)
		})
	}
}

func TestGenerateRecurringExpensesUseCase_RejectsBlankHouseholdIDBeforeRepositoryAccess(t *testing.T) {
	t.Parallel()

	for _, householdID := range []string{"", " \t "} {
		t.Run("household_id="+householdID, func(t *testing.T) {
			repo := &authorizationRecurringExpenseRepo{}
			members := &authorizationMemberRepo{}
			_, err := NewGenerateRecurringExpensesUseCase(repo, members, &mockExpenseRepo{}, &mockIDGen{}).Execute(context.Background(), inbound.GenerateRecurringExpensesCommand{
				HouseholdID:   householdID,
				CurrentUserID: "user-1",
			})
			if !errors.Is(err, recurringexpense.ErrInvalidHouseholdID) {
				t.Fatalf("error = %v; want ErrInvalidHouseholdID", err)
			}
			if repo.listDueForGenerationCalls != 0 {
				t.Fatalf("ListDueForGeneration calls = %d; want 0", repo.listDueForGenerationCalls)
			}
			if members.calls != 0 {
				t.Fatalf("FindByUserID calls = %d; want 0", members.calls)
			}
		})
	}
}
