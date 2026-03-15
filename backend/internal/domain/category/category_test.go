package category_test

import (
	"testing"
	"time"

	"micha/backend/internal/domain/category"
)

func TestNewCategory_Success(t *testing.T) {
	t.Parallel()
	c, err := category.New("cat-1", "hh-1", "Gym", "gym", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Slug() != "gym" {
		t.Errorf("Slug = %q; want %q", c.Slug(), "gym")
	}
	if c.IsDefault() {
		t.Error("custom category should not be default")
	}
}

func TestNewCategory_InvalidSlugCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		slug string
	}{
		{"empty", ""},
		{"spaces", "my category"},
		{"uppercase", "MyCategory"},
		{"special chars", "cat!"},
		{"leading hyphen", "-cat"},
		{"trailing hyphen", "cat-"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := category.New("id", "hh-1", "Test", tc.slug, time.Now())
			if err == nil {
				t.Errorf("expected error for slug %q, got nil", tc.slug)
			}
		})
	}
}

func TestNewCategory_ValidSlugCases(t *testing.T) {
	t.Parallel()
	cases := []string{"rent", "auto", "my-category", "cat1", "a-b-c"}
	for _, slug := range cases {
		slug := slug
		t.Run(slug, func(t *testing.T) {
			t.Parallel()
			_, err := category.New("id", "hh-1", "Test", slug, time.Now())
			if err != nil {
				t.Errorf("unexpected error for slug %q: %v", slug, err)
			}
		})
	}
}

func TestNewCategory_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := category.New("id", "hh-1", "  ", "gym", time.Now())
	if err == nil {
		t.Error("expected error for empty name")
	}
}
