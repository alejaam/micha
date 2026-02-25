package expenseapp_test

import (
	"context"
	"sync"

	"micha/backend/internal/domain/expense"
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
