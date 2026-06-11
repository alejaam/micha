package inbound

import (
	"context"
	"time"
)

// InitializePeriodInput defines data to create the very first period.
type InitializePeriodInput struct {
	HouseholdID   string
	CurrentUserID string
}

type InitializePeriodOutput struct {
	PeriodID string
}

type InitializePeriodUseCase interface {
	Execute(ctx context.Context, input InitializePeriodInput) (InitializePeriodOutput, error)
}

// SettlementEntry represents a single transfer between members in a settlement projection.
type SettlementEntry struct {
	FromMemberID string `json:"from_member_id"`
	ToMemberID   string `json:"to_member_id"`
	AmountCents  int64  `json:"amount_cents"`
}

// SimulateClosePeriodInput defines data to simulate closing a period (read-only).
type SimulateClosePeriodInput struct {
	HouseholdID   string
	PeriodID      string
	CurrentUserID string
}

// SimulateClosePeriodOutput contains the read-only projection of what would happen if the period closed.
type SimulateClosePeriodOutput struct {
	NextPeriodStart   time.Time         `json:"next_period_start"`
	NextPeriodEnd     time.Time         `json:"next_period_end"`
	SettlementPreview []SettlementEntry `json:"settlement_preview"`
	FixedExpenseCount int               `json:"fixed_expense_count"`
	InstallmentCount  int               `json:"installment_count"`
}

// SimulateClosePeriodUseCase contract for read-only period closure simulation.
type SimulateClosePeriodUseCase interface {
	Execute(ctx context.Context, input SimulateClosePeriodInput) (SimulateClosePeriodOutput, error)
}
