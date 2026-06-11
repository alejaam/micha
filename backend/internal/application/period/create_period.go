package periodapp

import (
	"context"
	"fmt"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/period"
	"micha/backend/internal/ports/outbound"
)

// InitialPeriodResult contains the outcome of CreateInitialPeriod.
type InitialPeriodResult struct {
	PeriodID  string
	StartDate time.Time
	EndDate   time.Time
}

// CreateInitialPeriod creates the initial period for a household based on its config.
// This is a shared helper usable from both InitializePeriodUseCase and RegisterHouseholdUseCase.
func CreateInitialPeriod(
	ctx context.Context,
	periodRepo outbound.PeriodRepository,
	idGen appshared.IDGenerator,
	householdID string,
	closingDay int,
	frequency string,
	now time.Time,
) (InitialPeriodResult, error) {
	var start, end time.Time
	if frequency == "biweekly" {
		if now.Day() <= 15 {
			start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			end = time.Date(now.Year(), now.Month(), 15, 23, 59, 59, 999999999, now.Location())
		} else {
			start = time.Date(now.Year(), now.Month(), 16, 0, 0, 0, 0, now.Location())
			end = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, now.Location())
		}
	} else {
		if now.Day() <= closingDay {
			lastMonth := now.AddDate(0, -1, 0)
			start = time.Date(lastMonth.Year(), lastMonth.Month(), closingDay+1, 0, 0, 0, 0, now.Location())
			end = time.Date(now.Year(), now.Month(), closingDay, 23, 59, 59, 999999999, now.Location())
		} else {
			start = time.Date(now.Year(), now.Month(), closingDay+1, 0, 0, 0, 0, now.Location())
			nextMonth := now.AddDate(0, 1, 0)
			end = time.Date(nextMonth.Year(), nextMonth.Month(), closingDay, 23, 59, 59, 999999999, now.Location())
		}
	}

	p, err := period.New(
		period.ID(idGen.NewID()),
		householdID,
		start,
		end,
		period.StatusOpen,
		now,
	)
	if err != nil {
		return InitialPeriodResult{}, fmt.Errorf("create initial period: %w", err)
	}

	if err := periodRepo.Create(ctx, p); err != nil {
		return InitialPeriodResult{}, fmt.Errorf("create initial period: %w", err)
	}

	return InitialPeriodResult{
		PeriodID:  string(p.ID()),
		StartDate: start,
		EndDate:   end,
	}, nil
}
