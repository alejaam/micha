package memberapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	memberapp "micha/backend/internal/application/member"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/invitation"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

type staticMemberIDGen string

func (s staticMemberIDGen) NewID() string { return string(s) }

type mockMemberRepo struct {
	members map[string]member.Member
	saveErr error
	listErr error
}

type mockInviteRepo struct {
	invites []invitation.Invitation
}

func (m *mockInviteRepo) Save(_ context.Context, inv invitation.Invitation) error {
	m.invites = append(m.invites, inv)
	return nil
}

type mockInviteSender struct {
	sentEmail string
	sentCode  string
}

func (m *mockInviteSender) SendInviteCode(_ context.Context, email, code string) error {
	m.sentEmail = email
	m.sentCode = code
	return nil
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[string]member.Member)}
}

func (m *mockMemberRepo) Save(_ context.Context, item member.Member) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.members[string(item.ID())] = item
	return nil
}

func (m *mockMemberRepo) FindByID(_ context.Context, id string) (member.Member, error) {
	item, ok := m.members[id]
	if !ok {
		return member.Member{}, errors.New("not found")
	}
	return item, nil
}

func (m *mockMemberRepo) ListByHousehold(_ context.Context, householdID string, limit, offset int) ([]member.Member, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]member.Member, 0, len(m.members))
	for _, item := range m.members {
		if item.HouseholdID() == householdID {
			result = append(result, item)
		}
	}
	if offset >= len(result) {
		return []member.Member{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockMemberRepo) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	for _, item := range m.members {
		if item.HouseholdID() == householdID && item.UserID() == userID {
			return item, nil
		}
	}
	return member.Member{}, errors.New("not found")
}

func (m *mockMemberRepo) Update(_ context.Context, item member.Member) error {
	m.members[string(item.ID())] = item
	return nil
}

func (m *mockMemberRepo) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]member.Member, 0, len(m.members))
	for _, item := range m.members {
		if item.HouseholdID() == householdID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockMemberRepo) FindByUserIDGlobal(_ context.Context, userID string) (member.Member, error) {
	for _, item := range m.members {
		if item.UserID() == userID {
			return item, nil
		}
	}
	return member.Member{}, errors.New("not found")
}

func (m *mockMemberRepo) ListHouseholdIDsByUserID(_ context.Context, userID string) ([]string, error) {
	var ids []string
	for _, item := range m.members {
		if item.UserID() == userID {
			ids = append(ids, item.HouseholdID())
		}
	}
	return ids, nil
}

func (m *mockMemberRepo) Delete(_ context.Context, id string) error {
	delete(m.members, id)
	return nil
}

func (m *mockMemberRepo) CountActiveByHousehold(_ context.Context, householdID string) (int, error) {
	count := 0
	for _, item := range m.members {
		if item.HouseholdID() == householdID {
			count++
		}
	}
	return count, nil
}

func TestRegisterMember_Success(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	uc := memberapp.NewRegisterMemberUseCase(repo, staticMemberIDGen("m-1"))

	out, err := uc.Execute(context.Background(), inbound.RegisterMemberInput{
		HouseholdID:        "hh-1",
		Name:               "Ale",
		Email:              "ale@mail.com",
		MonthlySalaryCents: 200000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.MemberID != "m-1" {
		t.Errorf("MemberID = %q; want %q", out.MemberID, "m-1")
	}
}

func TestRegisterMember_InvalidEmail(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	uc := memberapp.NewRegisterMemberUseCase(repo, staticMemberIDGen("m-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterMemberInput{
		HouseholdID:        "hh-1",
		Name:               "Ale",
		Email:              "ale.mail.com",
		MonthlySalaryCents: 200000,
	})
	if !errors.Is(err, member.ErrInvalidEmail) {
		t.Errorf("want ErrInvalidEmail, got %v", err)
	}
}

func TestListMembers_Success(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	now := time.Now()
	m, _ := member.New(member.ID("m-1"), "hh-1", "Ale", "ale@mail.com", 100, now)
	_ = repo.Save(context.Background(), m)
	uc := memberapp.NewListMembersUseCase(repo)

	items, err := uc.Execute(context.Background(), inbound.ListMembersQuery{HouseholdID: "hh-1", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d; want 1", len(items))
	}
}

func TestRegisterMember_OnlyOwnerCanCreateAfterBootstrap(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	now := time.Now()
	owner, _ := member.NewWithUserID(member.ID("m-owner"), "hh-1", "Owner", "owner@mail.com", "u-owner", 0, now)
	nonOwner, _ := member.NewWithUserID(member.ID("m-member"), "hh-1", "Member", "member@mail.com", "u-member", 0, now.Add(time.Minute))
	repo.members[string(owner.ID())] = owner
	repo.members[string(nonOwner.ID())] = nonOwner

	uc := memberapp.NewRegisterMemberUseCase(repo, staticMemberIDGen("m-3"))

	_, err := uc.Execute(context.Background(), inbound.RegisterMemberInput{
		HouseholdID:        "hh-1",
		Name:               "Invitado",
		Email:              "inv@mail.com",
		MonthlySalaryCents: 0,
		CallerMemberID:     "m-member",
	})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("expected shared.ErrForbidden, got %v", err)
	}
}

func TestRegisterMember_PendingCreatesAndSendsInviteCode(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	inviteRepo := &mockInviteRepo{}
	sender := &mockInviteSender{}
	uc := memberapp.NewRegisterMemberUseCaseWithInvites(repo, staticMemberIDGen("m-1"), inviteRepo, sender)

	_, err := uc.Execute(context.Background(), inbound.RegisterMemberInput{
		HouseholdID:        "hh-1",
		Name:               "Ale",
		Email:              "ale@mail.com",
		MonthlySalaryCents: 200000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inviteRepo.invites) != 1 {
		t.Fatalf("expected 1 invitation, got %d", len(inviteRepo.invites))
	}
	if sender.sentEmail != "ale@mail.com" {
		t.Fatalf("expected invite email ale@mail.com, got %s", sender.sentEmail)
	}
	if sender.sentCode == "" {
		t.Fatal("expected invite code to be generated")
	}
}

type remainingSalaryHouseholdRepo struct {
	item household.Household
}

func (r remainingSalaryHouseholdRepo) Save(context.Context, household.Household) error { return nil }
func (r remainingSalaryHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	if string(r.item.ID()) != id {
		return household.Household{}, shared.ErrNotFound
	}
	return r.item, nil
}
func (r remainingSalaryHouseholdRepo) List(context.Context, int, int) ([]household.Household, error) {
	return nil, nil
}
func (r remainingSalaryHouseholdRepo) ListByUserID(context.Context, string, int, int) ([]household.Household, error) {
	return nil, nil
}
func (r remainingSalaryHouseholdRepo) Update(context.Context, household.Household) error { return nil }

type remainingSalaryExpenseRepo struct {
	expenses       []expense.Expense
	personalByUser int64
}

func (r remainingSalaryExpenseRepo) Save(context.Context, expense.Expense) error { return nil }
func (r remainingSalaryExpenseRepo) FindByID(context.Context, string) (expense.Expense, error) {
	return expense.Expense{}, shared.ErrNotFound
}
func (r remainingSalaryExpenseRepo) List(context.Context, string, int, int) ([]expense.Expense, error) {
	return nil, nil
}
func (r remainingSalaryExpenseRepo) ListByHouseholdAndPeriod(_ context.Context, householdID string, from, to time.Time) ([]expense.Expense, error) {
	result := make([]expense.Expense, 0)
	for _, e := range r.expenses {
		if e.HouseholdID() == householdID && !e.CreatedAt().Before(from) && e.CreatedAt().Before(to) {
			result = append(result, e)
		}
	}
	return result, nil
}
func (r remainingSalaryExpenseRepo) SumPersonalByMemberAndPeriod(context.Context, string, string, time.Time, time.Time) (int64, error) {
	return r.personalByUser, nil
}
func (r remainingSalaryExpenseRepo) Update(context.Context, expense.Expense) error { return nil }

type remainingSalaryInstallmentRepo struct{}

func (r remainingSalaryInstallmentRepo) Save(context.Context, installment.Installment) error {
	return nil
}
func (r remainingSalaryInstallmentRepo) SaveAll(context.Context, []installment.Installment) error {
	return nil
}
func (r remainingSalaryInstallmentRepo) ListByExpense(context.Context, string) ([]installment.Installment, error) {
	return nil, nil
}
func (r remainingSalaryInstallmentRepo) ListByHouseholdAndPeriod(context.Context, string, time.Time, time.Time) ([]installment.Installment, error) {
	return nil, nil
}
func (r remainingSalaryInstallmentRepo) DeleteByExpense(context.Context, string) error { return nil }

func TestCalculateRemainingSalary_Success(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 4, 12, 0, 0, 0, time.UTC)
	memberRepo := newMockMemberRepo()
	m1, _ := member.NewWithUserID(member.ID("m-1"), "hh-1", "Ale", "ale@mail.com", "u-1", 100000, now)
	m2, _ := member.NewWithUserID(member.ID("m-2"), "hh-1", "Bea", "bea@mail.com", "u-2", 100000, now)
	memberRepo.members[string(m1.ID())] = m1
	memberRepo.members[string(m2.ID())] = m2

	hh, _ := household.NewFromAttributes(household.Attributes{
		ID:             household.ID("hh-1"),
		Name:           "Hogar",
		SettlementMode: household.SettlementModeEqual,
		Currency:       "MXN",
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	hhRepo := remainingSalaryHouseholdRepo{item: hh}

	sharedExpense, _ := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID("e-1"),
		HouseholdID:    "hh-1",
		PaidByMemberID: "m-1",
		AmountCents:    20000,
		Description:    "Super",
		IsShared:       true,
		Currency:       "MXN",
		PaymentMethod:  expense.PaymentMethodCash,
		ExpenseType:    expense.ExpenseTypeVariable,
		CategoryID:     "cat-other",
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	expenseRepo := remainingSalaryExpenseRepo{expenses: []expense.Expense{sharedExpense}, personalByUser: 15000}

	uc := memberapp.NewCalculateRemainingSalaryUseCase(hhRepo, memberRepo, expenseRepo, remainingSalaryInstallmentRepo{})
	out, err := uc.Execute(context.Background(), inbound.CalculateRemainingSalaryInput{
		HouseholdID: "hh-1",
		MemberID:    "m-1",
		Year:        2026,
		Month:       3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.MonthlySalaryCents != 100000 {
		t.Fatalf("MonthlySalaryCents = %d; want 100000", out.MonthlySalaryCents)
	}
	if out.PersonalExpensesCents != 15000 {
		t.Fatalf("PersonalExpensesCents = %d; want 15000", out.PersonalExpensesCents)
	}
	if out.AllocatedDebtCents != 10000 {
		t.Fatalf("AllocatedDebtCents = %d; want 10000", out.AllocatedDebtCents)
	}
	if out.RemainingSalaryCents != 75000 {
		t.Fatalf("RemainingSalaryCents = %d; want 75000", out.RemainingSalaryCents)
	}
}
