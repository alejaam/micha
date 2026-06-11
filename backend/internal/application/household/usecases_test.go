package householdapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	householdapp "micha/backend/internal/application/household"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/ports/inbound"
)

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
