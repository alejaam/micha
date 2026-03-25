package expenseapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	expenseapp "micha/backend/internal/application/expense"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// --- RegisterExpenseUseCase -------------------------------------------------

func TestRegisterExpense_Success(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	mRepo.seedMember("m-1", "hh-1")
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, catRepo, staticIDGen("exp-1"))

	out, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		AmountCents:    1500,
		Description:    "Taxi",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "variable",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ExpenseID != "exp-1" {
		t.Errorf("ExpenseID = %q; want %q", out.ExpenseID, "exp-1")
	}
}

func TestRegisterExpense_InvalidMoney(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	mRepo.seedMember("m-1", "hh-1")
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, catRepo, staticIDGen("exp-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		AmountCents:    0,
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "variable",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRegisterExpense_InvalidExpenseType(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	mRepo.seedMember("m-1", "hh-1")
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, catRepo, staticIDGen("exp-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		AmountCents:    1500,
		Description:    "Taxi",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "weird",
	})
	if !errors.Is(err, expense.ErrInvalidExpenseType) {
		t.Errorf("want ErrInvalidExpenseType, got %v", err)
	}
}

// --- GetExpenseUseCase ------------------------------------------------------

func TestGetExpense_Found(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	seedExpense(t, repo, "exp-1", "hh-1", 1000)
	uc := expenseapp.NewGetExpenseUseCase(repo)

	e, err := uc.Execute(context.Background(), "exp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(e.ID()) != "exp-1" {
		t.Errorf("ID = %q; want %q", e.ID(), "exp-1")
	}
}

func TestGetExpense_NotFound(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	uc := expenseapp.NewGetExpenseUseCase(repo)

	_, err := uc.Execute(context.Background(), "missing-id")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestGetExpense_SoftDeletedReturnsNotFound(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	e := seedExpense(t, repo, "exp-1", "hh-1", 1000)
	_ = e.SoftDelete()
	_ = repo.Update(context.Background(), *e)

	uc := expenseapp.NewGetExpenseUseCase(repo)
	_, err := uc.Execute(context.Background(), "exp-1")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("want ErrNotFound for deleted expense, got %v", err)
	}
}

// --- ListExpensesUseCase ----------------------------------------------------

func TestListExpenses_Success(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	seedExpense(t, repo, "exp-1", "hh-1", 1000)
	seedExpense(t, repo, "exp-2", "hh-1", 2000)
	uc := expenseapp.NewListExpensesUseCase(repo)

	expenses, err := uc.Execute(context.Background(), inbound.ListExpensesQuery{
		HouseholdID: "hh-1",
		Limit:       10,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expenses) != 2 {
		t.Errorf("got %d expenses; want 2", len(expenses))
	}
}

func TestListExpenses_MissingHouseholdID(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	uc := expenseapp.NewListExpensesUseCase(repo)

	_, err := uc.Execute(context.Background(), inbound.ListExpensesQuery{Limit: 10})
	if err == nil {
		t.Fatal("expected error for empty household_id")
	}
}

func TestListExpenses_DefaultLimit(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	uc := expenseapp.NewListExpensesUseCase(repo)

	_, err := uc.Execute(context.Background(), inbound.ListExpensesQuery{
		HouseholdID: "hh-1",
		Limit:       0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListExpenses_MaxLimitClamped(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	uc := expenseapp.NewListExpensesUseCase(repo)

	_, err := uc.Execute(context.Background(), inbound.ListExpensesQuery{
		HouseholdID: "hh-1",
		Limit:       9999,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- PatchExpenseUseCase ----------------------------------------------------

func TestPatchExpense_Success(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	seedExpense(t, repo, "exp-1", "hh-1", 1000)
	uc := expenseapp.NewPatchExpenseUseCase(repo)

	newAmt := int64(2000)
	e, err := uc.Execute(context.Background(), inbound.PatchExpenseCommand{
		ID:          "exp-1",
		AmountCents: &newAmt,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.AmountCents() != 2000 {
		t.Errorf("AmountCents = %d; want 2000", e.AmountCents())
	}
}

func TestPatchExpense_NotFound(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	uc := expenseapp.NewPatchExpenseUseCase(repo)

	_, err := uc.Execute(context.Background(), inbound.PatchExpenseCommand{ID: "missing"})
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestPatchExpense_InvalidAmount(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	seedExpense(t, repo, "exp-1", "hh-1", 1000)
	uc := expenseapp.NewPatchExpenseUseCase(repo)

	bad := int64(-1)
	_, err := uc.Execute(context.Background(), inbound.PatchExpenseCommand{
		ID:          "exp-1",
		AmountCents: &bad,
	})
	if !errors.Is(err, shared.ErrInvalidMoney) {
		t.Errorf("want ErrInvalidMoney, got %v", err)
	}
}

// --- DeleteExpenseUseCase ---------------------------------------------------

func TestDeleteExpense_Success(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	seedExpense(t, repo, "exp-1", "hh-1", 1000)
	uc := expenseapp.NewDeleteExpenseUseCase(repo)

	if err := uc.Execute(context.Background(), "exp-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e, _ := repo.FindByID(context.Background(), "exp-1")
	if e.DeletedAt() == nil {
		t.Error("expected DeletedAt to be set after delete")
	}
}

func TestDeleteExpense_NotFound(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	uc := expenseapp.NewDeleteExpenseUseCase(repo)

	err := uc.Execute(context.Background(), "missing")
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestDeleteExpense_AlreadyDeleted(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	e := seedExpense(t, repo, "exp-1", "hh-1", 1000)
	_ = e.SoftDelete()
	_ = repo.Update(context.Background(), *e)

	uc := expenseapp.NewDeleteExpenseUseCase(repo)
	err := uc.Execute(context.Background(), "exp-1")
	if !errors.Is(err, shared.ErrAlreadyDeleted) {
		t.Errorf("want ErrAlreadyDeleted, got %v", err)
	}
}

// --- helpers ----------------------------------------------------------------

type staticIDGen string

func (s staticIDGen) NewID() string { return string(s) }

func seedExpense(t *testing.T, repo *mockRepo, id, householdID string, amountCents int64) *expense.Expense {
	t.Helper()
	e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID(id),
		HouseholdID:    householdID,
		PaidByMemberID: "m-1",
		AmountCents:    amountCents,
		Description:    "seed",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  expense.PaymentMethodCash,
		ExpenseType:    expense.ExpenseTypeVariable,
		CreatedAt:      time.Now(),
	})
	if err != nil {
		t.Fatalf("seedExpense: %v", err)
	}
	if err := repo.Save(context.Background(), e); err != nil {
		t.Fatalf("seedExpense save: %v", err)
	}
	return &e
}
