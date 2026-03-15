package household

import (
	"errors"
)

var (
	// ErrInvalidSplitConfig is returned when member percentages don't sum to 100.
	ErrInvalidSplitConfig = errors.New("split percentages must sum to 100")
	// ErrEmptySplitConfig is returned when an empty slice is provided.
	ErrEmptySplitConfig = errors.New("split config must have at least one member")
)

// MemberSplit maps a member ID to a percentage (0–100, integer basis points not used here).
// Percentages are stored as integers (e.g. 60 means 60%).
type MemberSplit struct {
	MemberID   string
	Percentage int
}

// SplitConfig is a value object holding per-member split percentages for a household.
// An empty SplitConfig (nil or zero-length) means equal split among all members.
type SplitConfig struct {
	splits []MemberSplit
}

// NewSplitConfig validates and constructs a SplitConfig.
// Percentages must sum to exactly 100.
func NewSplitConfig(splits []MemberSplit) (SplitConfig, error) {
	if len(splits) == 0 {
		return SplitConfig{}, ErrEmptySplitConfig
	}

	total := 0
	for _, s := range splits {
		total += s.Percentage
	}
	if total != 100 {
		return SplitConfig{}, ErrInvalidSplitConfig
	}

	cp := make([]MemberSplit, len(splits))
	copy(cp, splits)
	return SplitConfig{splits: cp}, nil
}

// Splits returns the immutable list of member splits.
func (sc SplitConfig) Splits() []MemberSplit {
	cp := make([]MemberSplit, len(sc.splits))
	copy(cp, sc.splits)
	return cp
}

// IsEmpty reports whether this SplitConfig carries no data (equal-split default).
func (sc SplitConfig) IsEmpty() bool { return len(sc.splits) == 0 }
