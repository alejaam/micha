package householdapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// UpdateSplitConfigUseCase sets or replaces the per-member split percentages.
type UpdateSplitConfigUseCase struct {
	repo outbound.HouseholdRepository
}

// NewUpdateSplitConfigUseCase constructs an UpdateSplitConfigUseCase.
func NewUpdateSplitConfigUseCase(repo outbound.HouseholdRepository) UpdateSplitConfigUseCase {
	return UpdateSplitConfigUseCase{repo: repo}
}

func (u UpdateSplitConfigUseCase) Execute(ctx context.Context, input inbound.UpdateSplitConfigInput) error {
	h, err := u.repo.FindByID(ctx, input.HouseholdID)
	if err != nil {
		return fmt.Errorf("update split config: %w", err)
	}

	sc, err := household.NewSplitConfig(input.Splits)
	if err != nil {
		return fmt.Errorf("update split config: %w", err)
	}

	h.UpdateSplitConfig(sc)

	if err := u.repo.Update(ctx, h); err != nil {
		return fmt.Errorf("update split config: %w", err)
	}

	slog.InfoContext(ctx, "update split config", "household_id", input.HouseholdID)
	return nil
}

var _ inbound.UpdateSplitConfigUseCase = UpdateSplitConfigUseCase{}
