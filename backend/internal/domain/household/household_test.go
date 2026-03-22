package household_test

import (
	"errors"
	"testing"
	"time"

	"micha/backend/internal/domain/household"
)

func TestNewFromAttributes_Valid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	h, err := household.NewFromAttributes(household.Attributes{
		ID:             household.ID("hh-1"),
		Name:           "Casa",
		SettlementMode: household.SettlementModeProportional,
		Currency:       "mxn",
		CreatedAt:      now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Currency() != "MXN" {
		t.Errorf("Currency = %q; want %q", h.Currency(), "MXN")
	}
}

func TestNewFromAttributes_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		attrs household.Attributes
		want  error
	}{
		{
			name:  "empty name",
			attrs: household.Attributes{ID: "hh-1", Name: " ", SettlementMode: household.SettlementModeEqual, Currency: "MXN", CreatedAt: time.Now()},
			want:  household.ErrInvalidName,
		},
		{
			name:  "invalid settlement mode",
			attrs: household.Attributes{ID: "hh-1", Name: "Casa", SettlementMode: household.SettlementMode("random"), Currency: "MXN", CreatedAt: time.Now()},
			want:  household.ErrInvalidSettlementMode,
		},
		{
			name:  "invalid currency",
			attrs: household.Attributes{ID: "hh-1", Name: "Casa", SettlementMode: household.SettlementModeEqual, Currency: "MX", CreatedAt: time.Now()},
			want:  household.ErrInvalidCurrency,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := household.NewFromAttributes(tc.attrs)
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}
}
