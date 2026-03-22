package household

import "strings"

// MemberSplit represents how a member splits expenses within a household.
type MemberSplit struct {
	MemberID    string  // ID del miembro
	Percentage  float64 // Weight or percentage (0-100)
	Description string  // Optional: reason or display name
}

// SplitConfig represents the custom split configuration for a household.
// When empty, the household uses SettlementMode for default behavior.
type SplitConfig struct {
	splits []MemberSplit
}

// NewSplitConfig creates a new split configuration from member splits.
// Returns error if splits are invalid or empty.
func NewSplitConfig(splits []MemberSplit) (SplitConfig, error) {
	if len(splits) == 0 {
		return SplitConfig{}, ErrEmptySplitConfig
	}

	var total float64
	// Validate each split
	for _, s := range splits {
		if strings.TrimSpace(s.MemberID) == "" {
			return SplitConfig{}, ErrInvalidSplitConfig
		}
		if s.Percentage < 0 || s.Percentage > 100 {
			return SplitConfig{}, ErrInvalidSplitConfig
		}
		total += s.Percentage
	}

	if total != 100 {
		return SplitConfig{}, ErrInvalidSplitConfig
	}

	// Make a copy of the splits
	copiedSplits := make([]MemberSplit, len(splits))
	copy(copiedSplits, splits)

	return SplitConfig{splits: copiedSplits}, nil
}

// Splits returns a copy of all member splits.
func (sc SplitConfig) Splits() []MemberSplit {
	if len(sc.splits) == 0 {
		return []MemberSplit{}
	}
	copiedSplits := make([]MemberSplit, len(sc.splits))
	copy(copiedSplits, sc.splits)
	return copiedSplits
}

// IsEmpty returns true if the split config has no splits.
func (sc SplitConfig) IsEmpty() bool {
	return len(sc.splits) == 0
}
