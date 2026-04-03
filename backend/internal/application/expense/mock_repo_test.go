package expenseapp_test

import (
	"context"
	"sync"
	"time"

	"micha/backend/internal/domain/card"
	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/household"
	"micha/backend/internal/domain/installment"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/domain/shared"
)

type mockInstallmentRepo struct {
	mu           sync.RWMutex
	installments map[string]installment.Installment
}

func newMockInstallmentRepo() *mockInstallmentRepo {
	return &mockInstallmentRepo{installments: make(map[string]installment.Installment)}
}

func (m *mockInstallmentRepo) Save(_ context.Context, i installment.Installment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.installments[string(i.ID())] = i
	return nil
}

func (m *mockInstallmentRepo) SaveAll(_ context.Context, insts []installment.Installment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, i := range insts {
		m.installments[string(i.ID())] = i
	}
	return nil
}

func (m *mockInstallmentRepo) ListByExpense(_ context.Context, expenseID string) ([]installment.Installment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []installment.Installment
	for _, i := range m.installments {
		if i.ExpenseID() == expenseID {
			result = append(result, i)
		}
	}
	return result, nil
}

func (m *mockInstallmentRepo) ListByHouseholdAndPeriod(_ context.Context, _ string, _, _ time.Time) ([]installment.Installment, error) {
	return nil, nil
}

func (m *mockInstallmentRepo) DeleteByExpense(_ context.Context, expenseID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, i := range m.installments {
		if i.ExpenseID() == expenseID {
			delete(m.installments, id)
		}
	}
	return nil
}

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
func (r *mockHouseholdRepo) ListByUserID(_ context.Context, _ string, _, _ int) ([]household.Household, error) {
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
		UserID:      "user-linked",
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
func (r *mockMemberRepo) FindByUserIDGlobal(_ context.Context, _ string) (member.Member, error) {
	return member.Member{}, shared.ErrNotFound
}
func (r *mockMemberRepo) ListHouseholdIDsByUserID(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}
func (r *mockMemberRepo) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	var result []member.Member
	for _, m := range r.members {
		if m.HouseholdID() == householdID {
			result = append(result, m)
		}
	}
	return result, nil
}
func (r *mockMemberRepo) ListByHousehold(_ context.Context, _ string, _, _ int) ([]member.Member, error) {
	return nil, nil
}
func (r *mockMemberRepo) Update(_ context.Context, _ member.Member) error { return nil }
func (r *mockMemberRepo) Delete(_ context.Context, _ string) error        { return nil }
func (r *mockMemberRepo) CountActiveByHousehold(_ context.Context, _ string) (int, error) {
	return len(r.members), nil
}

type mockCategoryRepoActual struct {
	mu   sync.Mutex
	rows map[string]category.Category
}

type mockCardRepo struct {
	mu      sync.RWMutex
	cards   map[string]card.Card
	findErr error
}

func newMockCardRepo() *mockCardRepo {
	return &mockCardRepo{cards: make(map[string]card.Card)}
}

func (r *mockCardRepo) seedCard(id, householdID, cardName string) {
	c, _ := card.NewFromAttributes(card.Attributes{
		ID:          card.ID(id),
		HouseholdID: householdID,
		BankName:    "BANAMEX",
		CardName:    cardName,
		CutoffDay:   10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	})
	r.cards[id] = c
}

func (r *mockCardRepo) Save(_ context.Context, c card.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cards[string(c.ID())] = c
	return nil
}

func (r *mockCardRepo) FindByID(_ context.Context, id string) (card.Card, error) {
	if r.findErr != nil {
		return card.Card{}, r.findErr
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cards[id]
	if !ok {
		return card.Card{}, shared.ErrNotFound
	}
	return c, nil
}

func (r *mockCardRepo) ListByHousehold(_ context.Context, householdID string) ([]card.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]card.Card, 0)
	for _, c := range r.cards {
		if c.HouseholdID() == householdID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *mockCardRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.cards[id]; !ok {
		return shared.ErrNotFound
	}
	delete(r.cards, id)
	return nil
}

func newMockCategoryRepo() *mockCategoryRepoActual {
	return &mockCategoryRepoActual{rows: make(map[string]category.Category)}
}

func (r *mockCategoryRepoActual) Save(_ context.Context, c category.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[string(c.ID())] = c
	return nil
}

func (r *mockCategoryRepoActual) FindBySlug(_ context.Context, householdID, slug string) (category.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.rows {
		if c.HouseholdID() == householdID && c.Slug() == slug {
			return c, nil
		}
	}
	return category.Category{}, shared.ErrNotFound
}

func (r *mockCategoryRepoActual) ListByHousehold(_ context.Context, householdID string) ([]category.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []category.Category
	for _, c := range r.rows {
		if c.HouseholdID() == householdID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *mockCategoryRepoActual) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rows, id)
	return nil
}

func (r *mockCategoryRepoActual) seedCategory(id, householdID, slug string) {
	c, _ := category.NewFromAttributes(category.Attributes{
		ID:          category.ID(id),
		HouseholdID: householdID,
		Name:        slug, // simple for test
		Slug:        slug,
		CreatedAt:   time.Now(),
	})
	_ = r.Save(context.Background(), c)
}
