package member_test

import (
	"errors"
	"testing"
	"time"

	"micha/backend/internal/domain/member"
)

func TestNewFromAttributes_Valid(t *testing.T) {
	t.Parallel()
	m, err := member.NewFromAttributes(member.Attributes{
		ID:                 member.ID("m-1"),
		HouseholdID:        "hh-1",
		Name:               "Ale",
		Email:              "ALE@MAIL.COM",
		MonthlySalaryCents: 300000,
		CreatedAt:          time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Email() != "ale@mail.com" {
		t.Errorf("Email = %q; want %q", m.Email(), "ale@mail.com")
	}
}

func TestNewFromAttributes_Invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		attrs member.Attributes
		want  error
	}{
		{
			name:  "empty name",
			attrs: member.Attributes{ID: "m-1", HouseholdID: "hh-1", Name: " ", Email: "a@mail.com", MonthlySalaryCents: 100, CreatedAt: time.Now()},
			want:  member.ErrInvalidName,
		},
		{
			name:  "invalid email",
			attrs: member.Attributes{ID: "m-1", HouseholdID: "hh-1", Name: "Ale", Email: "ale.mail.com", MonthlySalaryCents: 100, CreatedAt: time.Now()},
			want:  member.ErrInvalidEmail,
		},
		{
			name:  "negative salary",
			attrs: member.Attributes{ID: "m-1", HouseholdID: "hh-1", Name: "Ale", Email: "ale@mail.com", MonthlySalaryCents: -1, CreatedAt: time.Now()},
			want:  member.ErrInvalidSalary,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := member.NewFromAttributes(tc.attrs)
			if !errors.Is(err, tc.want) {
				t.Errorf("want %v, got %v", tc.want, err)
			}
		})
	}
}
