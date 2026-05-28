package householdapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	householdapp "micha/backend/internal/application/household"
	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/inbound"
)

// --- Mocks ---

type staticRegIDGen string

func (s staticRegIDGen) NewID() string { return string(s) }

type mockRegHouseholdRepo struct {
	households map[string]household.Household
	saveErr    error
}

func newMockRegHouseholdRepo() *mockRegHouseholdRepo {
	return &mockRegHouseholdRepo{households: make(map[string]household.Household)}
}

func (m *mockRegHouseholdRepo) Save(_ context.Context, h household.Household) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.households[string(h.ID())] = h
	return nil
}

func (m *mockRegHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	h, ok := m.households[id]
	if !ok {
		return household.Household{}, errors.New("not found")
	}
	return h, nil
}

func (m *mockRegHouseholdRepo) List(_ context.Context, _, _ int) ([]household.Household, error) {
	return nil, nil
}

func (m *mockRegHouseholdRepo) ListByUserID(_ context.Context, _ string, _, _ int) ([]household.Household, error) {
	return nil, nil
}

func (m *mockRegHouseholdRepo) Update(_ context.Context, _ household.Household) error { return nil }

type mockRegCategoryRepo struct {
	categories []category.Category
}

func newMockRegCategoryRepo() *mockRegCategoryRepo {
	return &mockRegCategoryRepo{}
}

func (m *mockRegCategoryRepo) Save(_ context.Context, c category.Category) error {
	m.categories = append(m.categories, c)
	return nil
}

func (m *mockRegCategoryRepo) FindBySlug(_ context.Context, _, _ string) (category.Category, error) {
	return category.Category{}, errors.New("not found")
}

func (m *mockRegCategoryRepo) ListByHousehold(_ context.Context, _ string) ([]category.Category, error) {
	return m.categories, nil
}

func (m *mockRegCategoryRepo) Delete(_ context.Context, _ string) error { return nil }

type mockRegMemberRepo struct {
	members []member.Member
	saveErr error
}

func newMockRegMemberRepo() *mockRegMemberRepo {
	return &mockRegMemberRepo{}
}

func (m *mockRegMemberRepo) Save(_ context.Context, mem member.Member) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.members = append(m.members, mem)
	return nil
}

func (m *mockRegMemberRepo) FindByID(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, errors.New("not found")
}

func (m *mockRegMemberRepo) FindByUserID(_ context.Context, _, _ string) (member.Member, error) {
	return member.Member{}, errors.New("not found")
}

func (m *mockRegMemberRepo) FindByUserIDGlobal(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, errors.New("not found")
}

func (m *mockRegMemberRepo) ListAllByHousehold(_ context.Context, _ string) ([]member.Member, error) {
	return nil, nil
}

func (m *mockRegMemberRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]member.Member, error) {
	return nil, nil
}

func (m *mockRegMemberRepo) ListHouseholdIDsByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

func (m *mockRegMemberRepo) Update(_ context.Context, _ member.Member) error { return nil }
func (m *mockRegMemberRepo) Delete(_ context.Context, _ string) error        { return nil }
func (m *mockRegMemberRepo) CountActiveByHousehold(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *mockRegMemberRepo) LinkByEmail(_ context.Context, _, _ string) error { return nil }

type mockRegUserRepo struct {
	users map[string]*user.User
}

func newMockRegUserRepo(email string) *mockRegUserRepo {
	u, err := user.NewFromAttributes(user.UserAttributes{
		ID:           "u-1",
		Email:        email,
		PasswordHash: "hashed-password",
		CreatedAt:    time.Now(),
	})
	if err != nil {
		panic("failed to create test user: " + err.Error())
	}
	return &mockRegUserRepo{
		users: map[string]*user.User{
			"user-123": &u,
		},
	}
}

func (m *mockRegUserRepo) Save(_ context.Context, _ user.User) error { return nil }
func (m *mockRegUserRepo) FindByEmail(_ context.Context, _ string) (user.User, error) {
	for _, u := range m.users {
		return *u, nil
	}
	return user.User{}, errors.New("not found")
}

func (m *mockRegUserRepo) FindByID(_ context.Context, id string) (user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return user.User{}, errors.New("not found")
	}
	return *u, nil
}

type mockRegPeriodRepo struct {
	periods []period.Period
}

func newMockRegPeriodRepo() *mockRegPeriodRepo {
	return &mockRegPeriodRepo{}
}

func (m *mockRegPeriodRepo) Create(_ context.Context, p period.Period) error {
	m.periods = append(m.periods, p)
	return nil
}

func (m *mockRegPeriodRepo) GetByID(_ context.Context, _ period.ID) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockRegPeriodRepo) Update(_ context.Context, _ period.Period) error { return nil }

func (m *mockRegPeriodRepo) GetCurrentOpen(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockRegPeriodRepo) GetLatestByHousehold(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockRegPeriodRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]period.Period, error) {
	return nil, nil
}

type mockRegTxManager struct {
	runFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockRegTxManager) Run(_ context.Context, fn func(ctx context.Context) error) error {
	if m.runFn != nil {
		return m.runFn(context.Background(), fn)
	}
	return fn(context.Background())
}

// --- Tests ---

func TestRegisterHousehold_Success(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	householdRepo := newMockRegHouseholdRepo()
	categoryRepo := newMockRegCategoryRepo()
	memberRepo := newMockRegMemberRepo()
	userRepo := newMockRegUserRepo("test@example.com")
	periodRepo := newMockRegPeriodRepo()
	txManager := &mockRegTxManager{}

	uc := householdapp.NewRegisterHouseholdUseCase(
		householdRepo, categoryRepo, memberRepo, userRepo,
		periodRepo, txManager, staticRegIDGen("hh-1"),
	)
	uc.Now = func() time.Time { return now }

	out, err := uc.Execute(context.Background(), inbound.RegisterHouseholdInput{
		Name:             "Casa",
		SettlementMode:   household.SettlementModeEqual,
		Currency:         "MXN",
		ClosingDay:       15,
		PeriodFrequency:  "biweekly",
		CurrentUserID:    "user-123",
		OwnerSalaryCents: 5000000, // $50,000.00
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.HouseholdID != "hh-1" {
		t.Errorf("HouseholdID = %q; want %q", out.HouseholdID, "hh-1")
	}

	if out.MemberID == "" {
		t.Error("expected non-empty MemberID")
	}

	if out.PeriodID == "" {
		t.Error("expected non-empty PeriodID")
	}

	// Verify owner member was created with salary
	if len(memberRepo.members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(memberRepo.members))
	}
	if memberRepo.members[0].MonthlySalaryCents() != 5000000 {
		t.Errorf("MonthlySalaryCents = %d; want 5000000", memberRepo.members[0].MonthlySalaryCents())
	}
	if memberRepo.members[0].UserID() != "user-123" {
		t.Errorf("UserID = %q; want user-123", memberRepo.members[0].UserID())
	}

	// Verify period was created
	if len(periodRepo.periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(periodRepo.periods))
	}
	if periodRepo.periods[0].HouseholdID() != "hh-1" {
		t.Errorf("period HouseholdID = %q; want hh-1", periodRepo.periods[0].HouseholdID())
	}

	// Verify categories were seeded
	if len(categoryRepo.categories) == 0 {
		t.Error("expected categories to be seeded")
	}
}

func TestRegisterHousehold_InvalidName(t *testing.T) {
	t.Parallel()

	householdRepo := newMockRegHouseholdRepo()
	categoryRepo := newMockRegCategoryRepo()
	memberRepo := newMockRegMemberRepo()
	userRepo := newMockRegUserRepo("test@example.com")
	periodRepo := newMockRegPeriodRepo()
	txManager := &mockRegTxManager{}

	uc := householdapp.NewRegisterHouseholdUseCase(
		householdRepo, categoryRepo, memberRepo, userRepo,
		periodRepo, txManager, staticRegIDGen("hh-1"),
	)

	_, err := uc.Execute(context.Background(), inbound.RegisterHouseholdInput{
		Name:            " ",
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		CurrentUserID:   "user-123",
		OwnerSalaryCents: 0,
	})
	if !errors.Is(err, household.ErrInvalidName) {
		t.Errorf("want ErrInvalidName, got %v", err)
	}
}

func TestRegisterHousehold_NegativeSalary(t *testing.T) {
	t.Parallel()

	householdRepo := newMockRegHouseholdRepo()
	categoryRepo := newMockRegCategoryRepo()
	memberRepo := newMockRegMemberRepo()
	userRepo := newMockRegUserRepo("test@example.com")
	periodRepo := newMockRegPeriodRepo()
	txManager := &mockRegTxManager{}

	uc := householdapp.NewRegisterHouseholdUseCase(
		householdRepo, categoryRepo, memberRepo, userRepo,
		periodRepo, txManager, staticRegIDGen("hh-1"),
	)

	_, err := uc.Execute(context.Background(), inbound.RegisterHouseholdInput{
		Name:             "Casa",
		SettlementMode:   household.SettlementModeEqual,
		Currency:         "MXN",
		CurrentUserID:    "user-123",
		OwnerSalaryCents: -100,
	})
	if err == nil {
		t.Fatal("expected error for negative salary, got nil")
	}
}
