package period

import (
	"errors"
	"testing"
	"time"
)

func TestNew_PeriodTooShort(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC) // Only 4 days

	_, err := New(
		ID("per-1"),
		"hh-1",
		start,
		end,
		StatusOpen,
		time.Now(),
	)
	if err == nil {
		t.Fatal("expected error for period shorter than 7 days, got nil")
	}
	if !errors.Is(err, ErrPeriodTooShort) {
		t.Errorf("expected ErrPeriodTooShort, got: %v", err)
	}
}

func TestNew_PeriodExactlySevenDays(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC) // Exactly 7 days

	p, err := New(
		ID("per-1"),
		"hh-1",
		start,
		end,
		StatusOpen,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("expected no error for 7-day period, got: %v", err)
	}
	if p.ID() != ID("per-1") {
		t.Errorf("expected ID per-1, got: %s", p.ID())
	}
}

func TestNew_PeriodLongerThanSevenDays(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC) // 30 days

	p, err := New(
		ID("per-1"),
		"hh-1",
		start,
		end,
		StatusOpen,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("expected no error for 30-day period, got: %v", err)
	}
	if p.ID() != ID("per-1") {
		t.Errorf("expected ID per-1, got: %s", p.ID())
	}
}

func TestNewFromAttributes_ShortPeriodRejected(t *testing.T) {
	t.Parallel()

	_, err := NewFromAttributes(PeriodAttributes{
		ID:          ID("per-1"),
		HouseholdID: "hh-1",
		StartDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 1, 6, 23, 59, 59, 0, time.UTC), // Just under 7 days
		Status:      StatusOpen,
		CreatedAt:   time.Now(),
	})
	if err == nil {
		t.Fatal("expected error for period shorter than 7 days, got nil")
	}
	if !errors.Is(err, ErrPeriodTooShort) {
		t.Errorf("expected ErrPeriodTooShort, got: %v", err)
	}
}

func TestNewFromAttributes_SixDayPeriodRejected(t *testing.T) {
	t.Parallel()

	_, err := NewFromAttributes(PeriodAttributes{
		ID:          ID("per-1"),
		HouseholdID: "hh-1",
		StartDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC), // 6 days exactly
		Status:      StatusOpen,
		CreatedAt:   time.Now(),
	})
	if err == nil {
		t.Fatal("expected error for 6-day period, got nil")
	}
	if !errors.Is(err, ErrPeriodTooShort) {
		t.Errorf("expected ErrPeriodTooShort, got: %v", err)
	}
}
