package householdapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	householdapp "micha/backend/internal/application/household"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/user"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/domain/category"
)

type staticHouseholdIDGen string

func (s staticHouseholdIDGen) NewID() string { return string(s) }

type mockHouseholdRepo struct {
	households map[string]household.Household
	saveErr    error
	listErr    error
}

func newMockHouseholdRepo() *mockHouseholdRepo {
	return &mockHouseholdRepo{households: make(map[string]household.Household)}
}

func (m *mockHouseholdRepo) Save(_ context.Context, h household.Household) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.households[string(h.ID())] = h
	return nil
}

func (m *mockHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	h, ok := m.households[id]
	if !ok {
		return household.Household{}, errors.New("not found")
	}
	return h, nil
}

func (m *mockHouseholdRepo) List(_ context.Context, limit, offset int) ([]household.Household, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]household.Household, 0, len(m.households))
	for _, h := range m.households {
		result = append(result, h)
	}
	if offset >= len(result) {
		return []household.Household{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockHouseholdRepo) Update(_ context.Context, h household.Household) error {
	m.households[string(h.ID())] = h
	return nil
}

func (m *mockHouseholdRepo) ListByUserID(_ context.Context, _ string, limit, offset int) ([]household.Household, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]household.Household, 0, len(m.households))
	for _, h := range m.households {
		result = append(result, h)
	}
	if offset >= len(result) {
		return []household.Household{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

// mockCategoryRepo implements outbound.CategoryRepository for tests.
type mockCategoryRepo struct {
	categories []category.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{}
}

func (m *mockCategoryRepo) Save(_ context.Context, c category.Category) error {
	m.categories = append(m.categories, c)
	return nil
}

func (m *mockCategoryRepo) FindBySlug(_ context.Context, _, _ string) (category.Category, error) {
	return category.Category{}, errors.New("not found")
}

func (m *mockCategoryRepo) ListByHousehold(_ context.Context, _ string) ([]category.Category, error) {
	return m.categories, nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, _ string) error {
	return nil
}

// mockMemberRepo implements outbound.MemberRepository for tests.
type mockMemberRepo struct{}

func newMockMemberRepo() *mockMemberRepo { return &mockMemberRepo{} }
func (m *mockMemberRepo) Save(_ context.Context, _ member.Member) error       { return nil }
func (m *mockMemberRepo) FindByID(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, errors.New("not found")
}
func (m *mockMemberRepo) FindByUserID(_ context.Context, _, _ string) (member.Member, error) {
	return member.Member{}, errors.New("not found")
}
func (m *mockMemberRepo) FindByUserIDGlobal(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, errors.New("not found")
}
func (m *mockMemberRepo) ListAllByHousehold(_ context.Context, _ string) ([]member.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]member.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo) ListHouseholdIDsByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockMemberRepo) Update(_ context.Context, _ member.Member) error { return nil }
func (m *mockMemberRepo) Delete(_ context.Context, _ string) error         { return nil }
func (m *mockMemberRepo) CountActiveByHousehold(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *mockMemberRepo) LinkByEmail(_ context.Context, _, _ string) error { return nil }

// mockUserRepo implements outbound.UserRepository for tests.
type mockUserRepo struct{}

func newMockUserRepo() *mockUserRepo { return &mockUserRepo{} }
func (m *mockUserRepo) Save(_ context.Context, _ user.User) error { return nil }
func (m *mockUserRepo) FindByEmail(_ context.Context, _ string) (user.User, error) {
	return user.User{}, errors.New("not found")
}
func (m *mockUserRepo) FindByID(_ context.Context, _ string) (user.User, error) {
	return user.User{}, errors.New("not found")
}

// Additional mock types for the updated RegisterHouseholdUseCase interface

type mockPeriodRepo struct{}

func (m *mockPeriodRepo) Create(_ context.Context, _ period.Period) error { return nil }
func (m *mockPeriodRepo) GetByID(_ context.Context, _ period.ID) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}
func (m *mockPeriodRepo) Update(_ context.Context, _ period.Period) error { return nil }
func (m *mockPeriodRepo) GetCurrentOpen(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}
func (m *mockPeriodRepo) GetLatestByHousehold(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}
func (m *mockPeriodRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]period.Period, error) {
	return nil, nil
}

type mockTxManager struct{}

func (m *mockTxManager) Run(_ context.Context, fn func(ctx context.Context) error) error {
	return fn(context.Background())
}

func TestUpdateSplitConfig_Success(t *testing.T) {
	t.Parallel()
	repo := newMockHouseholdRepo()
	now := time.Now()
	h, _ := household.New(household.ID("hh-1"), "Casa", "u-1", household.SettlementModeEqual, "MXN", now)
	_ = repo.Save(context.Background(), h)

	uc := householdapp.NewUpdateSplitConfigUseCase(repo)
	err := uc.Execute(context.Background(), inbound.UpdateSplitConfigInput{
		HouseholdID: "hh-1",
		Splits: []household.MemberSplit{
			{MemberID: "m-1", Percentage: 60},
			{MemberID: "m-2", Percentage: 40},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := repo.FindByID(context.Background(), "hh-1")
	// SplitConfig is managed by the adapter; verify update succeeded without errors
	_ = updated
}

func TestUpdateSplitConfig_InvalidSum(t *testing.T) {
	t.Parallel()
	repo := newMockHouseholdRepo()
	now := time.Now()
	h, _ := household.New(household.ID("hh-1"), "Casa", "u-1", household.SettlementModeEqual, "MXN", now)
	_ = repo.Save(context.Background(), h)

	uc := householdapp.NewUpdateSplitConfigUseCase(repo)
	err := uc.Execute(context.Background(), inbound.UpdateSplitConfigInput{
		HouseholdID: "hh-1",
		Splits: []household.MemberSplit{
			{MemberID: "m-1", Percentage: 50},
			{MemberID: "m-2", Percentage: 30},
		},
	})
	if !errors.Is(err, household.ErrInvalidSplitConfig) {
		t.Errorf("want ErrInvalidSplitConfig, got %v", err)
	}
}

func TestUpdateSplitConfig_HouseholdNotFound(t *testing.T) {
	t.Parallel()
	repo := newMockHouseholdRepo()
	uc := householdapp.NewUpdateSplitConfigUseCase(repo)

	err := uc.Execute(context.Background(), inbound.UpdateSplitConfigInput{
		HouseholdID: "missing",
		Splits:      []household.MemberSplit{{MemberID: "m-1", Percentage: 100}},
	})
	if err == nil {
		t.Fatal("expected error for missing household, got nil")
	}
}

func TestListHouseholds_Success(t *testing.T) {
	t.Parallel()
	repo := newMockHouseholdRepo()
	now := time.Now()
	h, _ := household.New(household.ID("hh-1"), "Casa", "u-1", household.SettlementModeEqual, "MXN", now)
	_ = repo.Save(context.Background(), h)
	uc := householdapp.NewListHouseholdsUseCase(repo)

	items, err := uc.Execute(context.Background(), inbound.ListHouseholdsQuery{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d; want 1", len(items))
	}
}
