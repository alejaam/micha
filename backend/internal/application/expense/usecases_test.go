package expenseapp_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	expenseapp "micha/backend/internal/application/expense"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/member"
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
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"))

	out, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		CurrentUserID:  "user-linked",
		AmountCents:    1500,
		Description:    "Taxi",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "fixed",
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
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		CurrentUserID:  "user-linked",
		AmountCents:    0,
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "fixed",
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
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		CurrentUserID:  "user-linked",
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

func TestRegisterExpense_MSI_GeneratesInstallments(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	now := time.Now()
	mRepo.seedMemberWithUser("m-owner", "hh-1", "u-owner", now)
	mRepo.seedMemberWithUser("m-1", "hh-1", "user-linked", now.Add(time.Minute))
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	// Use a sequential ID generator to avoid collisions between root expense and installments
	seqIDGen := &sequentialIDGen{prefix: "exp-"}
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, seqIDGen)

	out, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:       "hh-1",
		PaidByMemberID:    "m-1",
		CurrentUserID:     "user-linked",
		AmountCents:       1000,
		Description:       "iPhone",
		IsShared:          true,
		Currency:          "MXN",
		PaymentMethod:     "card",
		ExpenseType:       "msi",
		TotalInstallments: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	installments, _ := instRepo.ListByExpense(context.Background(), out.ExpenseID)
	if len(installments) != 3 {
		t.Errorf("got %d installments; want 3", len(installments))
	}

	// The first installment should be 334 (333 + 1 remainder)
	foundFirst := false
	for _, i := range installments {
		if i.CurrentInstallment() == 1 {
			foundFirst = true
			if i.InstallmentAmountCents() != 334 {
				t.Errorf("inst[1] amount = %d; want 334", i.InstallmentAmountCents())
			}
		}
	}
	if !foundFirst {
		t.Error("installment 1 not found")
	}
}

type sequentialIDGen struct {
	mu      sync.Mutex
	prefix  string
	counter int
}

func (s *sequentialIDGen) NewID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	return fmt.Sprintf("%s%d", s.prefix, s.counter)
}

func TestRegisterExpense_PendingMember_Rejected(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := &pendingMemberMock{mockMemberRepo: *newMockMemberRepo()}
	mRepo.seedMember("m-pending", "hh-1")
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-pending",
		CurrentUserID:  "user-linked",
		AmountCents:    1000,
		ExpenseType:    "variable",
	})
	if err == nil {
		t.Fatal("expected pending member rejection error")
	}
}

func TestRegisterExpense_WithCardID_UsesRegisteredCardName(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	mRepo.seedMember("m-1", "hh-1")
	cardRepo := newMockCardRepo()
	cardRepo.seedCard("card-1", "hh-1", "Banamex Oro")
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCase(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		CurrentUserID:  "user-linked",
		AmountCents:    2200,
		Description:    "Uber",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "card",
		ExpenseType:    "fixed",
		CardID:         "card-1",
		CardName:       "Manual Name",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e, findErr := repo.FindByID(context.Background(), "exp-1")
	if findErr != nil {
		t.Fatalf("find expense: %v", findErr)
	}

	if e.CardID() != "card-1" {
		t.Errorf("CardID = %q; want %q", e.CardID(), "card-1")
	}
	if e.CardName() != "Banamex Oro" {
		t.Errorf("CardName = %q; want %q", e.CardName(), "Banamex Oro")
	}
}

func TestRegisterExpense_OwnerOnlyFixed(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	now := time.Now()
	mRepo.seedMemberWithUser("m-owner", "hh-1", "u-owner", now)
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCaseWithPolicy(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"), true)

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-owner",
		CurrentUserID:  "u-owner",
		AmountCents:    1200,
		Description:    "test",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "msi",
	})
	if !errors.Is(err, expenseapp.ErrExpenseTypeNotAllowedByRole) {
		t.Fatalf("expected ErrExpenseTypeNotAllowedByRole, got %v", err)
	}
}

func TestRegisterExpense_MemberCannotCreateFixed(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	now := time.Now()
	mRepo.seedMemberWithUser("m-owner", "hh-1", "u-owner", now)
	mRepo.seedMemberWithUser("m-member", "hh-1", "u-member", now.Add(time.Minute))
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCaseWithPolicy(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"), true)

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-member",
		CurrentUserID:  "u-member",
		AmountCents:    1200,
		Description:    "test",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "fixed",
	})
	if !errors.Is(err, expenseapp.ErrExpenseTypeNotAllowedByRole) {
		t.Fatalf("expected ErrExpenseTypeNotAllowedByRole, got %v", err)
	}
}

func TestRegisterExpense_OwnerOnBehalfControlledByFlag(t *testing.T) {
	t.Parallel()
	repo := newMockRepo()
	hhRepo := newMockHouseholdRepo("hh-1")
	mRepo := newMockMemberRepo()
	now := time.Now()
	mRepo.seedMemberWithUser("m-owner", "hh-1", "u-owner", now)
	mRepo.seedMemberWithUser("m-member", "hh-1", "u-member", now.Add(time.Minute))
	cardRepo := newMockCardRepo()
	catRepo := newMockCategoryRepo()
	catRepo.seedCategory("cat-other", "hh-1", "other")
	instRepo := newMockInstallmentRepo()
	uc := expenseapp.NewRegisterExpenseUseCaseWithPolicy(repo, hhRepo, mRepo, cardRepo, catRepo, instRepo, staticIDGen("exp-1"), false)

	_, err := uc.Execute(context.Background(), inbound.RegisterExpenseInput{
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-member",
		CurrentUserID:  "u-owner",
		AmountCents:    1200,
		Description:    "test",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  "cash",
		ExpenseType:    "fixed",
	})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("expected shared.ErrForbidden, got %v", err)
	}
}

type pendingMemberMock struct {
	mockMemberRepo
}

func (r *pendingMemberMock) seedMember(id, householdID string) {
	m, _ := member.New(member.ID(id), householdID, "Pending", "p@mail.com", 0, time.Now())
	r.members[id] = m
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
