package expenseapp_test

import (
	"context"
	"sync"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
)

// mockRepo is a hand-written in-memory mock for outbound.ExpenseRepository.
type mockRepo struct {
	mu       sync.RWMutex
	expenses map[string]expense.Expense

	saveErr   error
	findErr   error
	listErr   error
	updateErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{expenses: make(map[string]expense.Expense)}
}

func (m *mockRepo) Save(_ context.Context, e expense.Expense) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expenses[string(e.ID())] = e
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id string) (expense.Expense, error) {
	if m.findErr != nil {
		return expense.Expense{}, m.findErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.expenses[id]
	if !ok {
		return expense.Expense{}, shared.ErrNotFound
	}
	return e, nil
}

func (m *mockRepo) List(_ context.Context, householdID string, limit, offset int) ([]expense.Expense, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []expense.Expense
	for _, e := range m.expenses {
		if e.HouseholdID() == householdID && e.DeletedAt() == nil {
			result = append(result, e)
		}
	}
	if offset >= len(result) {
		return []expense.Expense{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockRepo) Update(_ context.Context, e expense.Expense) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.expenses[string(e.ID())]; !ok {
		return shared.ErrNotFound
	}
	m.expenses[string(e.ID())] = e
	return nil
}

func (m *mockRepo) ListByHouseholdAndPeriod(_ context.Context, householdID string, from, to time.Time) ([]expense.Expense, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]expense.Expense, 0)
	for _, e := range m.expenses {
		if e.HouseholdID() != householdID || e.DeletedAt() != nil {
			continue
		}
		if e.CreatedAt().Before(from) || !e.CreatedAt().Before(to) {
			continue
		}
		result = append(result, e)
	}

	return result, nil
}

// mockHouseholdRepo is a minimal in-memory mock for outbound.HouseholdRepository.
type mockHouseholdRepo struct {
	households map[string]household.Household
	findErr    error
}

func newMockHouseholdRepo(ids ...string) *mockHouseholdRepo {
	r := &mockHouseholdRepo{households: make(map[string]household.Household)}
	for _, id := range ids {
		h, _ := household.NewFromAttributes(household.Attributes{
			ID:             household.ID(id),
			Name:           "test household",
			SettlementMode: household.SettlementModeEqual,
			Currency:       "MXN",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})
		r.households[id] = h
	}
	return r
}

func (r *mockHouseholdRepo) FindByID(_ context.Context, id string) (household.Household, error) {
	if r.findErr != nil {
		return household.Household{}, r.findErr
	}
	h, ok := r.households[id]
	if !ok {
		return household.Household{}, shared.ErrNotFound
	}
	return h, nil
}

func (r *mockHouseholdRepo) Save(_ context.Context, _ household.Household) error { return nil }
func (r *mockHouseholdRepo) List(_ context.Context, _, _ int) ([]household.Household, error) {
	return nil, nil
}
func (r *mockHouseholdRepo) Update(_ context.Context, _ household.Household) error { return nil }

// mockMemberRepo is a minimal in-memory mock for outbound.MemberRepository.
type mockMemberRepo struct {
	members map[string]member.Member
	findErr error
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[string]member.Member)}
}

func (r *mockMemberRepo) seedMember(id, householdID string) {
	m, _ := member.NewFromAttributes(member.Attributes{
		ID:          member.ID(id),
		HouseholdID: householdID,
		Name:        "Test Member",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	r.members[id] = m
}

func (r *mockMemberRepo) FindByID(_ context.Context, id string) (member.Member, error) {
	if r.findErr != nil {
		return member.Member{}, r.findErr
	}
	m, ok := r.members[id]
	if !ok {
		return member.Member{}, shared.ErrNotFound
	}
	return m, nil
}

func (r *mockMemberRepo) Save(_ context.Context, _ member.Member) error { return nil }
func (r *mockMemberRepo) FindByUserID(_ context.Context, _, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (r *mockMemberRepo) ListAllByHousehold(_ context.Context, _ string) ([]member.Member, error) {
	return nil, nil
}
func (r *mockMemberRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]member.Member, error) {
	return nil, nil
}
func (r *mockMemberRepo) Update(_ context.Context, _ member.Member) error { return nil }
