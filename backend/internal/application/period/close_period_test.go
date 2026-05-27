package periodapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/period"
	periodapproval "micha/backend/internal/domain/period_approval"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockPeriodRepo struct {
	periods map[period.ID]period.Period
}

func newMockPeriodRepo() *mockPeriodRepo {
	return &mockPeriodRepo{periods: make(map[period.ID]period.Period)}
}

func (m *mockPeriodRepo) Create(_ context.Context, p period.Period) error {
	m.periods[p.ID()] = p
	return nil
}

func (m *mockPeriodRepo) GetByID(_ context.Context, id period.ID) (period.Period, error) {
	p, ok := m.periods[id]
	if !ok {
		return period.Period{}, shared.ErrNotFound
	}
	return p, nil
}

func (m *mockPeriodRepo) Update(_ context.Context, p period.Period) error {
	m.periods[p.ID()] = p
	return nil
}

func (m *mockPeriodRepo) GetCurrentOpen(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockPeriodRepo) GetLatestByHousehold(_ context.Context, _ string) (period.Period, error) {
	return period.Period{}, shared.ErrNotFound
}

func (m *mockPeriodRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]period.Period, error) {
	return nil, nil
}

type mockHouseholdRepo struct {
	households map[string]household.Household
}

func newMockHouseholdRepo() *mockHouseholdRepo {
	return &mockHouseholdRepo{households: make(map[string]household.Household)}
}

func (m *mockHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	h, ok := m.households[id]
	if !ok {
		return household.Household{}, shared.ErrNotFound
	}
	return h, nil
}

// Unused interface methods
func (m *mockHouseholdRepo) Save(_ context.Context, _ household.Household) error { return nil }
func (m *mockHouseholdRepo) List(_ context.Context, _, _ int) ([]household.Household, error) {
	return nil, nil
}
func (m *mockHouseholdRepo) ListByUserID(_ context.Context, _ string, _, _ int) ([]household.Household, error) {
	return nil, nil
}
func (m *mockHouseholdRepo) Update(_ context.Context, _ household.Household) error { return nil }

type mockMemberRepo2 struct {
	members []member.Member
}

func (m *mockMemberRepo2) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID && mem.UserID() == userID {
			return mem, nil
		}
	}
	return member.Member{}, shared.ErrNotFound
}

func (m *mockMemberRepo2) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	var result []member.Member
	for _, mem := range m.members {
		if mem.HouseholdID() == householdID {
			result = append(result, mem)
		}
	}
	return result, nil
}

// Unused interface methods
func (m *mockMemberRepo2) Save(_ context.Context, _ member.Member) error { return nil }
func (m *mockMemberRepo2) FindByID(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *mockMemberRepo2) FindByUserIDGlobal(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (m *mockMemberRepo2) ListByHousehold(_ context.Context, _ string, _, _ int) ([]member.Member, error) {
	return nil, nil
}
func (m *mockMemberRepo2) ListHouseholdIDsByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (m *mockMemberRepo2) Update(_ context.Context, _ member.Member) error { return nil }
func (m *mockMemberRepo2) Delete(_ context.Context, _ string) error        { return nil }
func (m *mockMemberRepo2) CountActiveByHousehold(_ context.Context, _ string) (int, error) {
	return 0, nil
}
func (m *mockMemberRepo2) LinkByEmail(_ context.Context, _, _ string) error { return nil }

type mockApprovalRepo struct {
	approvals []periodapproval.PeriodApproval
}

func (m *mockApprovalRepo) ListByPeriod(_ context.Context, periodID string) ([]periodapproval.PeriodApproval, error) {
	var result []periodapproval.PeriodApproval
	for _, a := range m.approvals {
		if a.PeriodID() == periodID {
			result = append(result, a)
		}
	}
	return result, nil
}

// Unused interface methods
func (m *mockApprovalRepo) Save(_ context.Context, _ periodapproval.PeriodApproval) error {
	return nil
}
func (m *mockApprovalRepo) GetByMemberAndPeriod(_ context.Context, _, _ string) (periodapproval.PeriodApproval, error) {
	return periodapproval.PeriodApproval{}, shared.ErrNotFound
}
func (m *mockApprovalRepo) DeleteAllByPeriod(_ context.Context, _ string) error { return nil }

type mockExpenseRepo struct {
	expenses []expense.Expense
}

func (m *mockExpenseRepo) ListByPeriod(_ context.Context, _ string) ([]expense.Expense, error) {
	return m.expenses, nil
}

func (m *mockExpenseRepo) FindByID(_ context.Context, id string) (expense.Expense, error) {
	for _, e := range m.expenses {
		if string(e.ID()) == id {
			return e, nil
		}
	}
	return expense.Expense{}, shared.ErrNotFound
}

func (m *mockExpenseRepo) Save(_ context.Context, _ expense.Expense) error { return nil }

// Unused interface methods
func (m *mockExpenseRepo) List(_ context.Context, _ string, _, _ int) ([]expense.Expense, error) {
	return nil, nil
}
func (m *mockExpenseRepo) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]expense.Expense, error) {
	return nil, nil
}
func (m *mockExpenseRepo) SumPersonalByMemberAndPeriod(_ context.Context, _, _ string, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *mockExpenseRepo) Update(_ context.Context, _ expense.Expense) error { return nil }
func (m *mockExpenseRepo) AdoptOrphanExpenses(_ context.Context, _, _ string, _, _ time.Time) error {
	return nil
}

type mockInstallmentRepo struct {
	installments []installment.Installment
}

func (m *mockInstallmentRepo) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]installment.Installment, error) {
	return m.installments, nil
}

// Unused interface methods
func (m *mockInstallmentRepo) Save(_ context.Context, _ installment.Installment) error { return nil }
func (m *mockInstallmentRepo) SaveAll(_ context.Context, _ []installment.Installment) error {
	return nil
}
func (m *mockInstallmentRepo) ListByExpense(_ context.Context, _ string) ([]installment.Installment, error) {
	return nil, nil
}
func (m *mockInstallmentRepo) DeleteByExpense(_ context.Context, _ string) error { return nil }

type mockTxManager struct{}

func (m *mockTxManager) Run(_ context.Context, fn func(ctx context.Context) error) error {
	return fn(context.Background())
}

type mockIDGen struct{ counter int }

func (m *mockIDGen) NewID() string {
	m.counter++
	return "id-" + string(rune('a'+m.counter-1))
}

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

func makeTestPeriodInReview(t *testing.T, id, householdID string) period.Period {
	t.Helper()
	p, err := period.NewFromAttributes(period.PeriodAttributes{
		ID:          period.ID(id),
		HouseholdID: householdID,
		StartDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
		Status:      period.StatusReview,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to create test period: %v", err)
	}
	return p
}

func makeTestHouseholdWithOwner(t *testing.T, id, ownerID string) household.Household {
	t.Helper()
	h, err := household.NewFromAttributes(household.Attributes{
		ID:              household.ID(id),
		Name:            "Test Household",
		OwnerID:         ownerID,
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
	return h
}

func makeTestMember(t *testing.T, id, householdID, userID string) member.Member {
	t.Helper()
	m, err := member.NewFromAttributes(member.Attributes{
		ID:          member.ID(id),
		HouseholdID: householdID,
		Name:        "Test Member",
		Email:       "test@example.com",
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to create test member: %v", err)
	}
	return m
}

func makeApproval(t *testing.T, memberID, periodID string, status periodapproval.ApprovalStatus) periodapproval.PeriodApproval {
	t.Helper()
	a, err := periodapproval.New(
		periodapproval.ID("a-"+memberID),
		memberID,
		periodID,
		status,
		"",
		time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to create test approval: %v", err)
	}
	return a
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestClosePeriodUseCase_OwnerCanCloseWithConsensus(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	ownerMemberID := "m-owner"

	periodRepo := newMockPeriodRepo()
	p := makeTestPeriodInReview(t, periodID, householdID)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = makeTestHouseholdWithOwner(t, householdID, ownerUserID)

	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, ownerMemberID, householdID, ownerUserID),
		},
	}

	approvalRepo := &mockApprovalRepo{
		approvals: []periodapproval.PeriodApproval{
			makeApproval(t, ownerMemberID, periodID, periodapproval.ApprovalStatusApproved),
		},
	}

	uc := NewClosePeriodUseCase(
		periodRepo,
		approvalRepo,
		householdRepo,
		memberRepo,
		&mockExpenseRepo{},
		&mockInstallmentRepo{},
		&mockTxManager{},
		&mockIDGen{},
	)
	// Override clock to avoid depending on real time
	uc.now = func() time.Time { return time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) }

	output, err := uc.Execute(context.Background(), inbound.ClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: ownerUserID,
		Force:         false,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if output.NextPeriodID == "" {
		t.Error("expected non-empty NextPeriodID")
	}
}

func TestClosePeriodUseCase_OwnerCanForceClose(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	ownerMemberID := "m-owner"

	periodRepo := newMockPeriodRepo()
	p := makeTestPeriodInReview(t, periodID, householdID)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = makeTestHouseholdWithOwner(t, householdID, ownerUserID)

	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, ownerMemberID, householdID, ownerUserID),
		},
	}

	// No approvals at all — force skips consensus
	approvalRepo := &mockApprovalRepo{}

	uc := NewClosePeriodUseCase(
		periodRepo,
		approvalRepo,
		householdRepo,
		memberRepo,
		&mockExpenseRepo{},
		&mockInstallmentRepo{},
		&mockTxManager{},
		&mockIDGen{},
	)
	uc.now = func() time.Time { return time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) }

	output, err := uc.Execute(context.Background(), inbound.ClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: ownerUserID,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if output.NextPeriodID == "" {
		t.Error("expected non-empty NextPeriodID")
	}
}

func TestClosePeriodUseCase_NonOwnerGetsForbidden(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	nonOwnerUserID := "user-non-owner"
	nonOwnerMemberID := "m-non-owner"

	periodRepo := newMockPeriodRepo()
	p := makeTestPeriodInReview(t, periodID, householdID)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = makeTestHouseholdWithOwner(t, householdID, ownerUserID)

	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, nonOwnerMemberID, householdID, nonOwnerUserID),
		},
	}

	approvalRepo := &mockApprovalRepo{}

	uc := NewClosePeriodUseCase(
		periodRepo,
		approvalRepo,
		householdRepo,
		memberRepo,
		&mockExpenseRepo{},
		&mockInstallmentRepo{},
		&mockTxManager{},
		&mockIDGen{},
	)
	uc.now = func() time.Time { return time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) }

	_, err := uc.Execute(context.Background(), inbound.ClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: nonOwnerUserID,
		Force:         false,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, shared.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestClosePeriodUseCase_NonOwnerWithConsensusStillGetsForbidden(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	nonOwnerUserID := "user-non-owner"
	nonOwnerMemberID := "m-non-owner"

	periodRepo := newMockPeriodRepo()
	p := makeTestPeriodInReview(t, periodID, householdID)
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = makeTestHouseholdWithOwner(t, householdID, ownerUserID)

	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, nonOwnerMemberID, householdID, nonOwnerUserID),
		},
	}

	// Even with full consensus, non-owner should be rejected
	approvalRepo := &mockApprovalRepo{
		approvals: []periodapproval.PeriodApproval{
			makeApproval(t, nonOwnerMemberID, periodID, periodapproval.ApprovalStatusApproved),
		},
	}

	uc := NewClosePeriodUseCase(
		periodRepo,
		approvalRepo,
		householdRepo,
		memberRepo,
		&mockExpenseRepo{},
		&mockInstallmentRepo{},
		&mockTxManager{},
		&mockIDGen{},
	)
	uc.now = func() time.Time { return time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC) }

	_, err := uc.Execute(context.Background(), inbound.ClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: nonOwnerUserID,
		Force:         false,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, shared.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got: %v", err)
	}
}

func TestClosePeriodUseCase_CannotCloseBeforeMinimumDuration(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-1"
	ownerUserID := "user-owner"
	ownerMemberID := "m-owner"

	// Period started 3 days ago — too short to close
	periodRepo := newMockPeriodRepo()
	p, err := period.NewFromAttributes(period.PeriodAttributes{
		ID:          period.ID(periodID),
		HouseholdID: householdID,
		StartDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
		Status:      period.StatusReview,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("failed to create test period: %v", err)
	}
	periodRepo.periods[p.ID()] = p

	householdRepo := newMockHouseholdRepo()
	householdRepo.households[householdID] = makeTestHouseholdWithOwner(t, householdID, ownerUserID)

	memberRepo := &mockMemberRepo2{
		members: []member.Member{
			makeTestMember(t, ownerMemberID, householdID, ownerUserID),
		},
	}

	// Full consensus
	approvalRepo := &mockApprovalRepo{
		approvals: []periodapproval.PeriodApproval{
			makeApproval(t, ownerMemberID, periodID, periodapproval.ApprovalStatusApproved),
		},
	}

	uc := NewClosePeriodUseCase(
		periodRepo,
		approvalRepo,
		householdRepo,
		memberRepo,
		&mockExpenseRepo{},
		&mockInstallmentRepo{},
		&mockTxManager{},
		&mockIDGen{},
	)
	// Today is only 3 days after start — should be rejected
	uc.now = func() time.Time { return time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC) }

	_, err = uc.Execute(context.Background(), inbound.ClosePeriodInput{
		HouseholdID:   householdID,
		PeriodID:      periodID,
		CurrentUserID: ownerUserID,
		Force:         false,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, period.ErrPeriodTooShort) {
		t.Errorf("expected ErrPeriodTooShort, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Future-period guard tests (SC-1, SC-2)
// ---------------------------------------------------------------------------

func TestClosePeriodUseCase_FuturePeriodGuard(t *testing.T) {
	t.Parallel()

	householdID := "hh-1"
	periodID := "per-future"
	ownerUserID := "user-owner"
	ownerMemberID := "m-owner"

	// Period ends Jan 31, so nextStart = Feb 1.
	setup := func(t *testing.T, now time.Time) ClosePeriodUseCase {
		t.Helper()

		periodRepo := newMockPeriodRepo()
		p := makeTestPeriodInReview(t, periodID, householdID)
		periodRepo.periods[p.ID()] = p

		householdRepo := newMockHouseholdRepo()
		householdRepo.households[householdID] = makeTestHouseholdWithOwner(t, householdID, ownerUserID)

		memberRepo := &mockMemberRepo2{
			members: []member.Member{
				makeTestMember(t, ownerMemberID, householdID, ownerUserID),
			},
		}

		approvalRepo := &mockApprovalRepo{
			approvals: []periodapproval.PeriodApproval{
				makeApproval(t, ownerMemberID, periodID, periodapproval.ApprovalStatusApproved),
			},
		}

		uc := NewClosePeriodUseCase(
			periodRepo,
			approvalRepo,
			householdRepo,
			memberRepo,
			&mockExpenseRepo{},
			&mockInstallmentRepo{},
			&mockTxManager{},
			&mockIDGen{},
		)
		uc.now = func() time.Time { return now }
		return uc
	}

	tests := []struct {
		name    string
		now     time.Time
		wantErr bool
		want    error // specific sentinel error expected
	}{
		{
			name:    "future period rejected when nextStart is after today",
			now:     time.Date(2026, 1, 31, 14, 37, 22, 0, time.UTC),
			wantErr: true,
			want:    shared.ErrFuturePeriod,
		},
		{
			name:    "boundary day allowed when nextStart equals today midnight",
			now:     time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
			want:    nil,
		},
		{
			name:    "time-component safety: non-zero clock time on boundary day",
			now:     time.Date(2026, 2, 1, 14, 37, 22, 0, time.UTC),
			wantErr: false,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := setup(t, tt.now)
			_, err := uc.Execute(context.Background(), inbound.ClosePeriodInput{
				HouseholdID:   householdID,
				PeriodID:      periodID,
				CurrentUserID: ownerUserID,
				Force:         true,
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.want) {
					t.Errorf("expected error %v, got: %v", tt.want, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}
