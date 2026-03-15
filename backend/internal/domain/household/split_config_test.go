package household_test

import (
	"testing"

	"micha/backend/internal/domain/household"
)

func TestNewSplitConfig_Success(t *testing.T) {
	t.Parallel()
	sc, err := household.NewSplitConfig([]household.MemberSplit{
		{MemberID: "m-1", Percentage: 60},
		{MemberID: "m-2", Percentage: 40},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sc.IsEmpty() {
		t.Error("expected non-empty split config")
	}
	if len(sc.Splits()) != 2 {
		t.Errorf("len(Splits) = %d; want 2", len(sc.Splits()))
	}
}

func TestNewSplitConfig_SumNot100(t *testing.T) {
	t.Parallel()
	cases := [][]household.MemberSplit{
		{{MemberID: "m-1", Percentage: 50}, {MemberID: "m-2", Percentage: 30}},
		{{MemberID: "m-1", Percentage: 100}, {MemberID: "m-2", Percentage: 1}},
		{{MemberID: "m-1", Percentage: 0}},
	}
	for _, splits := range cases {
		_, err := household.NewSplitConfig(splits)
		if err == nil {
			t.Errorf("expected error for splits %v, got nil", splits)
		}
	}
}

func TestNewSplitConfig_Empty(t *testing.T) {
	t.Parallel()
	_, err := household.NewSplitConfig(nil)
	if err == nil {
		t.Error("expected error for nil splits")
	}
}

func TestSplitConfig_Immutability(t *testing.T) {
	t.Parallel()
	sc, _ := household.NewSplitConfig([]household.MemberSplit{
		{MemberID: "m-1", Percentage: 100},
	})
	// Mutate the returned slice — original must be unaffected.
	got := sc.Splits()
	got[0] = household.MemberSplit{MemberID: "mutated", Percentage: 99}

	got2 := sc.Splits()
	if got2[0].MemberID == "mutated" {
		t.Error("SplitConfig is not immutable: internal slice was mutated")
	}
}
