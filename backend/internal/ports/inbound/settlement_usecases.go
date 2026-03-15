package inbound

import (
	"context"
	"errors"

	"micha/backend/internal/domain/household"
)

// Settlement validation sentinel errors.
var (
	ErrSettlementMissingHouseholdID = errors.New("household_id is required")
	ErrSettlementYearOutOfRange     = errors.New("year is out of range")
	ErrSettlementInvalidMonth       = errors.New("month must be between 1 and 12")
)

// CalculateSettlementInput contains report scope for monthly settlement.
type CalculateSettlementInput struct {
	HouseholdID string
	Year        int
	Month       int
}

// MemberSettlement is the per-member output in a settlement report.
type MemberSettlement struct {
	MemberID        string
	Name            string
	PaidCents       int64
	ExpectedShare   int64
	NetBalanceCents int64
	SalaryWeightBps int64
}

// SettlementTransfer is a recommended transfer to close balances.
type SettlementTransfer struct {
	FromMemberID string
	ToMemberID   string
	AmountCents  int64
}

// CalculateSettlementOutput is the monthly settlement report.
type CalculateSettlementOutput struct {
	HouseholdID             string
	Year                    int
	Month                   int
	SettlementMode          household.SettlementMode
	EffectiveSettlementMode household.SettlementMode
	FallbackReason          string
	TotalSharedCents        int64
	IncludedExpenseCount    int
	ExcludedVoucherCount    int
	Members                 []MemberSettlement
	Transfers               []SettlementTransfer
}

// CalculateSettlementUseCase calculates settlement for a household period.
type CalculateSettlementUseCase interface {
	Execute(ctx context.Context, input CalculateSettlementInput) (CalculateSettlementOutput, error)
}
