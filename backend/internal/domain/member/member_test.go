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

func TestNewWithUserID_LinksUser(t *testing.T) {
	t.Parallel()
	now := time.Now()
	m, err := member.NewWithUserID(member.ID("m-1"), "hh-1", "Ale", "ale@mail.com", "user-99", 100000, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.UserID() != "user-99" {
		t.Errorf("UserID = %q; want %q", m.UserID(), "user-99")
	}
}

func TestNewFromAttributes_UserIDPreserved(t *testing.T) {
	t.Parallel()
	m, err := member.NewFromAttributes(member.Attributes{
		ID:                 member.ID("m-1"),
		HouseholdID:        "hh-1",
		Name:               "Ale",
		Email:              "ale@mail.com",
		MonthlySalaryCents: 0,
		UserID:             "user-42",
		CreatedAt:          time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.UserID() != "user-42" {
		t.Errorf("UserID = %q; want %q", m.UserID(), "user-42")
	}
	attrs := m.Attributes()
	if attrs.UserID != "user-42" {
		t.Errorf("Attributes().UserID = %q; want %q", attrs.UserID, "user-42")
	}
}

func TestLinkUser_SetsUserID(t *testing.T) {
	t.Parallel()
	m, err := member.New(member.ID("m-1"), "hh-1", "Ale", "ale@mail.com", 0, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.UserID() != "" {
		t.Errorf("expected empty UserID before linking, got %q", m.UserID())
	}
	m.LinkUser("user-77")
	if m.UserID() != "user-77" {
		t.Errorf("UserID after LinkUser = %q; want %q", m.UserID(), "user-77")
	}
}
