package periodapp

import (
	"context"
	"testing"
	"time"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// ---------------------------------------------------------------------------
// InitializePeriodUseCase Tests
// ---------------------------------------------------------------------------

func TestInitializePeriodUseCase_Biweekly_FirstHalf(t *testing.T) {
	t.Parallel()

	householdID := "hh-biweekly-1"
	ownerUserID := "user-owner"

	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(householdID),
		Name:            "Test Household",
		OwnerID:         ownerUserID,
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		ClosingDay:      15,
		PeriodFrequency: "biweekly",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = h

	periodRepo := newMockPeriodRepo()
	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, "m-owner", householdID, ownerUserID),
		},
	}
	expenseRepo := &mockExpenseRepo{}

	uc := NewInitializePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		expenseRepo,
		&mockIDGen{},
	)
	// Today is May 5th — should create period May 1-15
	uc.now = func() time.Time { return time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC) }

	output, err := uc.Execute(context.Background(), inbound.InitializePeriodInput{
		HouseholdID:   householdID,
		CurrentUserID: ownerUserID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p, ok := periodRepo.periods[period.ID(output.PeriodID)]
	if !ok {
		t.Fatal("expected period to be created")
	}

	if p.StartDate().Day() != 1 {
		t.Errorf("expected start day 1, got %d", p.StartDate().Day())
	}
	if p.EndDate().Day() != 15 {
		t.Errorf("expected end day 15, got %d", p.EndDate().Day())
	}
	if p.StartDate().Month() != 5 {
		t.Errorf("expected start month May, got %v", p.StartDate().Month())
	}
}

func TestInitializePeriodUseCase_Biweekly_SecondHalf(t *testing.T) {
	t.Parallel()

	householdID := "hh-biweekly-2"
	ownerUserID := "user-owner"

	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(householdID),
		Name:            "Test Household",
		OwnerID:         ownerUserID,
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		ClosingDay:      15,
		PeriodFrequency: "biweekly",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = h

	periodRepo := newMockPeriodRepo()
	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, "m-owner", householdID, ownerUserID),
		},
	}
	expenseRepo := &mockExpenseRepo{}

	uc := NewInitializePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		expenseRepo,
		&mockIDGen{},
	)
	// Today is May 20th — should create period May 16-31
	uc.now = func() time.Time { return time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC) }

	output, err := uc.Execute(context.Background(), inbound.InitializePeriodInput{
		HouseholdID:   householdID,
		CurrentUserID: ownerUserID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p, ok := periodRepo.periods[period.ID(output.PeriodID)]
	if !ok {
		t.Fatal("expected period to be created")
	}

	if p.StartDate().Day() != 16 {
		t.Errorf("expected start day 16, got %d", p.StartDate().Day())
	}
	if p.EndDate().Day() != 31 {
		t.Errorf("expected end day 31, got %d", p.EndDate().Day())
	}
	if p.StartDate().Month() != 5 {
		t.Errorf("expected start month May, got %v", p.StartDate().Month())
	}
}

func TestInitializePeriodUseCase_Biweekly_FebruaryLeapYear(t *testing.T) {
	t.Parallel()

	householdID := "hh-biweekly-feb"
	ownerUserID := "user-owner"

	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(householdID),
		Name:            "Test Household",
		OwnerID:         ownerUserID,
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		ClosingDay:      15,
		PeriodFrequency: "biweekly",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = h

	periodRepo := newMockPeriodRepo()
	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, "m-owner", householdID, ownerUserID),
		},
	}
	expenseRepo := &mockExpenseRepo{}

	uc := NewInitializePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		expenseRepo,
		&mockIDGen{},
	)
	// Today is Feb 20th 2024 (leap year) — should create period Feb 16-29
	uc.now = func() time.Time { return time.Date(2024, 2, 20, 0, 0, 0, 0, time.UTC) }

	output, err := uc.Execute(context.Background(), inbound.InitializePeriodInput{
		HouseholdID:   householdID,
		CurrentUserID: ownerUserID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p, ok := periodRepo.periods[period.ID(output.PeriodID)]
	if !ok {
		t.Fatal("expected period to be created")
	}

	if p.StartDate().Day() != 16 {
		t.Errorf("expected start day 16, got %d", p.StartDate().Day())
	}
	if p.EndDate().Day() != 29 {
		t.Errorf("expected end day 29 (leap year), got %d", p.EndDate().Day())
	}
	if p.StartDate().Month() != 2 {
		t.Errorf("expected start month February, got %v", p.StartDate().Month())
	}
}

func TestInitializePeriodUseCase_Biweekly_FebruaryNonLeap(t *testing.T) {
	t.Parallel()

	householdID := "hh-biweekly-feb-nl"
	ownerUserID := "user-owner"

	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(householdID),
		Name:            "Test Household",
		OwnerID:         ownerUserID,
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		ClosingDay:      15,
		PeriodFrequency: "biweekly",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = h

	periodRepo := newMockPeriodRepo()
	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, "m-owner", householdID, ownerUserID),
		},
	}
	expenseRepo := &mockExpenseRepo{}

	uc := NewInitializePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		expenseRepo,
		&mockIDGen{},
	)
	// Today is Feb 20th 2025 (non-leap) — should create period Feb 16-28
	uc.now = func() time.Time { return time.Date(2025, 2, 20, 0, 0, 0, 0, time.UTC) }

	output, err := uc.Execute(context.Background(), inbound.InitializePeriodInput{
		HouseholdID:   householdID,
		CurrentUserID: ownerUserID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p, ok := periodRepo.periods[period.ID(output.PeriodID)]
	if !ok {
		t.Fatal("expected period to be created")
	}

	if p.StartDate().Day() != 16 {
		t.Errorf("expected start day 16, got %d", p.StartDate().Day())
	}
	if p.EndDate().Day() != 28 {
		t.Errorf("expected end day 28 (non-leap year), got %d", p.EndDate().Day())
	}
}

func TestInitializePeriodUseCase_Monthly(t *testing.T) {
	t.Parallel()

	householdID := "hh-monthly"
	ownerUserID := "user-owner"

	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(householdID),
		Name:            "Test Household",
		OwnerID:         ownerUserID,
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		ClosingDay:      15,
		PeriodFrequency: "monthly",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = h

	periodRepo := newMockPeriodRepo()
	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, "m-owner", householdID, ownerUserID),
		},
	}
	expenseRepo := &mockExpenseRepo{}

	uc := NewInitializePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		expenseRepo,
		&mockIDGen{},
	)
	// Today is May 10th — before closingDay, so period is Apr 16 - May 15
	uc.now = func() time.Time { return time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC) }

	output, err := uc.Execute(context.Background(), inbound.InitializePeriodInput{
		HouseholdID:   householdID,
		CurrentUserID: ownerUserID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	p, ok := periodRepo.periods[period.ID(output.PeriodID)]
	if !ok {
		t.Fatal("expected period to be created")
	}

	if p.StartDate().Day() != 16 {
		t.Errorf("expected start day 16, got %d", p.StartDate().Day())
	}
	if p.StartDate().Month() != 4 {
		t.Errorf("expected start month April, got %v", p.StartDate().Month())
	}
	if p.EndDate().Day() != 15 {
		t.Errorf("expected end day 15, got %d", p.EndDate().Day())
	}
	if p.EndDate().Month() != 5 {
		t.Errorf("expected end month May, got %v", p.EndDate().Month())
	}
}

// mockPeriodRepoWithLatest is a period repo that can return a latest period
type mockPeriodRepoWithLatest struct {
	*mockPeriodRepo
	latest period.Period
}

func (m *mockPeriodRepoWithLatest) GetLatestByHousehold(_ context.Context, _ string) (period.Period, error) {
	if m.latest.ID() == "" {
		return period.Period{}, shared.ErrNotFound
	}
	return m.latest, nil
}

func TestInitializePeriodUseCase_AlreadyHasPeriods(t *testing.T) {
	t.Parallel()

	householdID := "hh-existing"
	ownerUserID := "user-owner"

	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(householdID),
		Name:            "Test Household",
		OwnerID:         ownerUserID,
		SettlementMode:  household.SettlementModeEqual,
		Currency:        "MXN",
		ClosingDay:      15,
		PeriodFrequency: "biweekly",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = h

	existingPeriod, _ := period.NewFromAttributes(period.PeriodAttributes{
		ID:          period.ID("existing"),
		HouseholdID: householdID,
		StartDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		Status:      period.StatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})

	periodRepo := &mockPeriodRepoWithLatest{
		mockPeriodRepo: newMockPeriodRepo(),
		latest:         existingPeriod,
	}

	memberRepo := &mockMemberRepo2{}
	expenseRepo := &mockExpenseRepo{}

	uc := NewInitializePeriodUseCase(
		periodRepo,
		householdRepo,
		memberRepo,
		expenseRepo,
		&mockIDGen{},
	)
	uc.now = func() time.Time { return time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC) }

	_, err = uc.Execute(context.Background(), inbound.InitializePeriodInput{
		HouseholdID:   householdID,
		CurrentUserID: ownerUserID,
	})
	if err == nil {
		t.Fatal("expected error when household already has periods")
	}
}
