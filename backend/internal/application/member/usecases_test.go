package memberapp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	memberapp "micha/backend/internal/application/member"
	"micha/backend/internal/domain/member"
	"micha/backend/internal/ports/inbound"
)

type staticMemberIDGen string

func (s staticMemberIDGen) NewID() string { return string(s) }

type mockMemberRepo struct {
	members map[string]member.Member
	saveErr error
	listErr error
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{members: make(map[string]member.Member)}
}

func (m *mockMemberRepo) Save(_ context.Context, item member.Member) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.members[string(item.ID())] = item
	return nil
}

func (m *mockMemberRepo) FindByID(_ context.Context, id string) (member.Member, error) {
	item, ok := m.members[id]
	if !ok {
		return member.Member{}, errors.New("not found")
	}
	return item, nil
}

func (m *mockMemberRepo) ListByHousehold(_ context.Context, householdID string, limit, offset int) ([]member.Member, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]member.Member, 0, len(m.members))
	for _, item := range m.members {
		if item.HouseholdID() == householdID {
			result = append(result, item)
		}
	}
	if offset >= len(result) {
		return []member.Member{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockMemberRepo) FindByUserID(_ context.Context, householdID, userID string) (member.Member, error) {
	for _, item := range m.members {
		if item.HouseholdID() == householdID && item.UserID() == userID {
			return item, nil
		}
	}
	return member.Member{}, errors.New("not found")
}

func (m *mockMemberRepo) Update(_ context.Context, item member.Member) error {
	m.members[string(item.ID())] = item
	return nil
}

func (m *mockMemberRepo) ListAllByHousehold(_ context.Context, householdID string) ([]member.Member, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]member.Member, 0, len(m.members))
	for _, item := range m.members {
		if item.HouseholdID() == householdID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *mockMemberRepo) FindByUserIDGlobal(_ context.Context, userID string) (member.Member, error) {
	for _, item := range m.members {
		if item.UserID() == userID {
			return item, nil
		}
	}
	return member.Member{}, errors.New("not found")
}

func (m *mockMemberRepo) ListHouseholdIDsByUserID(_ context.Context, userID string) ([]string, error) {
	var ids []string
	for _, item := range m.members {
		if item.UserID() == userID {
			ids = append(ids, item.HouseholdID())
		}
	}
	return ids, nil
}

func (m *mockMemberRepo) Delete(_ context.Context, id string) error {
	delete(m.members, id)
	return nil
}

func (m *mockMemberRepo) CountActiveByHousehold(_ context.Context, householdID string) (int, error) {
	count := 0
	for _, item := range m.members {
		if item.HouseholdID() == householdID {
			count++
		}
	}
	return count, nil
}

func TestRegisterMember_Success(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	uc := memberapp.NewRegisterMemberUseCase(repo, staticMemberIDGen("m-1"))

	out, err := uc.Execute(context.Background(), inbound.RegisterMemberInput{
		HouseholdID:        "hh-1",
		Name:               "Ale",
		Email:              "ale@mail.com",
		MonthlySalaryCents: 200000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.MemberID != "m-1" {
		t.Errorf("MemberID = %q; want %q", out.MemberID, "m-1")
	}
}

func TestRegisterMember_InvalidEmail(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	uc := memberapp.NewRegisterMemberUseCase(repo, staticMemberIDGen("m-1"))

	_, err := uc.Execute(context.Background(), inbound.RegisterMemberInput{
		HouseholdID:        "hh-1",
		Name:               "Ale",
		Email:              "ale.mail.com",
		MonthlySalaryCents: 200000,
	})
	if !errors.Is(err, member.ErrInvalidEmail) {
		t.Errorf("want ErrInvalidEmail, got %v", err)
	}
}

func TestListMembers_Success(t *testing.T) {
	t.Parallel()
	repo := newMockMemberRepo()
	now := time.Now()
	m, _ := member.New(member.ID("m-1"), "hh-1", "Ale", "ale@mail.com", 100, now)
	_ = repo.Save(context.Background(), m)
	uc := memberapp.NewListMembersUseCase(repo)

	items, err := uc.Execute(context.Background(), inbound.ListMembersQuery{HouseholdID: "hh-1", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d; want 1", len(items))
	}
}
