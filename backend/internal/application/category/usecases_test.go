package categoryapp_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	categoryapp "micha/backend/internal/application/category"
	"micha/backend/internal/domain/category"
	"micha/backend/internal/domain/shared"
	"micha/backend/internal/ports/inbound"
)

// --- in-memory mock ---

type mockCategoryRepo struct {
	mu   sync.Mutex
	rows map[string]category.Category // id → Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{rows: make(map[string]category.Category)}
}

func (m *mockCategoryRepo) Save(_ context.Context, c category.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows[string(c.ID())] = c
	return nil
}

func (m *mockCategoryRepo) FindBySlug(_ context.Context, householdID, slug string) (category.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.rows {
		if c.HouseholdID() == householdID && c.Slug() == slug {
			return c, nil
		}
	}
	return category.Category{}, shared.ErrNotFound
}

func (m *mockCategoryRepo) ListByHousehold(_ context.Context, householdID string) ([]category.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []category.Category
	for _, c := range m.rows {
		if c.HouseholdID() == householdID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.rows[id]
	if !ok {
		return shared.ErrNotFound
	}
	if c.IsDefault() {
		return shared.ErrNotFound // default rows protected at DB level
	}
	delete(m.rows, id)
	return nil
}

// seed helper
func seedCategory(repo *mockCategoryRepo, id, householdID, name, slug string, isDefault bool) {
	_ = isDefault
	c, _ := category.NewFromAttributes(category.Attributes{
		ID:          category.ID(id),
		HouseholdID: householdID,
		Name:        name,
		Slug:        slug,
		CreatedAt:   time.Now(),
	})
	_ = repo.Save(context.Background(), c)
}

type staticIDGen string

func (s staticIDGen) NewID() string { return string(s) }

// --- CreateCategory tests ---

func TestCreateCategory_Success(t *testing.T) {
	t.Parallel()
	repo := newMockCategoryRepo()
	uc := categoryapp.NewCreateCategoryUseCase(repo, staticIDGen("cat-1"))

	out, err := uc.Execute(context.Background(), inbound.CreateCategoryInput{
		HouseholdID: "hh-1",
		Name:        "Gym",
		Slug:        "gym",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CategoryID != "cat-1" {
		t.Errorf("CategoryID = %q; want %q", out.CategoryID, "cat-1")
	}
}

func TestCreateCategory_DuplicateSlug(t *testing.T) {
	t.Parallel()
	repo := newMockCategoryRepo()
	seedCategory(repo, "existing", "hh-1", "Gym", "gym", false)
	uc := categoryapp.NewCreateCategoryUseCase(repo, staticIDGen("cat-2"))

	_, err := uc.Execute(context.Background(), inbound.CreateCategoryInput{
		HouseholdID: "hh-1",
		Name:        "Gym again",
		Slug:        "gym",
	})
	if !errors.Is(err, shared.ErrAlreadyExists) {
		t.Errorf("want ErrAlreadyExists, got %v", err)
	}
}

func TestCreateCategory_InvalidSlug(t *testing.T) {
	t.Parallel()
	repo := newMockCategoryRepo()
	uc := categoryapp.NewCreateCategoryUseCase(repo, staticIDGen("cat-3"))

	_, err := uc.Execute(context.Background(), inbound.CreateCategoryInput{
		HouseholdID: "hh-1",
		Name:        "Bad Slug",
		Slug:        "Bad Slug!",
	})
	if !errors.Is(err, category.ErrInvalidSlug) {
		t.Errorf("want ErrInvalidSlug, got %v", err)
	}
}

// --- ListCategories tests ---

func TestListCategories_ReturnsAll(t *testing.T) {
	t.Parallel()
	repo := newMockCategoryRepo()
	seedCategory(repo, "c1", "hh-1", "Rent", "rent", true)
	seedCategory(repo, "c2", "hh-1", "Gym", "gym", false)
	seedCategory(repo, "c3", "hh-2", "Food", "food", true)
	uc := categoryapp.NewListCategoriesUseCase(repo)

	cats, err := uc.Execute(context.Background(), inbound.ListCategoriesQuery{HouseholdID: "hh-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("got %d categories; want 2", len(cats))
	}
}

// --- DeleteCategory tests ---

func TestDeleteCategory_Success(t *testing.T) {
	t.Parallel()
	repo := newMockCategoryRepo()
	seedCategory(repo, "c1", "hh-1", "Gym", "gym", false)
	uc := categoryapp.NewDeleteCategoryUseCase(repo)

	if err := uc.Execute(context.Background(), inbound.DeleteCategoryInput{
		HouseholdID: "hh-1",
		CategoryID:  "c1",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cats, _ := repo.ListByHousehold(context.Background(), "hh-1")
	if len(cats) != 0 {
		t.Errorf("expected 0 categories after delete, got %d", len(cats))
	}
}

func TestDeleteCategory_DefaultForbidden(t *testing.T) {
	t.Parallel()
	repo := newMockCategoryRepo()
	seedCategory(repo, "c1", "hh-1", "Rent", "rent", true)
	uc := categoryapp.NewDeleteCategoryUseCase(repo)

	err := uc.Execute(context.Background(), inbound.DeleteCategoryInput{
		HouseholdID: "hh-1",
		CategoryID:  "c1",
	})
	if !errors.Is(err, categoryapp.ErrCannotDeleteDefault) {
		t.Errorf("want ErrCannotDeleteDefault, got %v", err)
	}
}

func TestDeleteCategory_NotFound(t *testing.T) {
	t.Parallel()
	repo := newMockCategoryRepo()
	uc := categoryapp.NewDeleteCategoryUseCase(repo)

	err := uc.Execute(context.Background(), inbound.DeleteCategoryInput{
		HouseholdID: "hh-1",
		CategoryID:  "missing",
	})
	if !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
