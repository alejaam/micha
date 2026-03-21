package householdapp

import (
	"context"
	"fmt"
	"log/slog"

	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

var _ inbound.UpdateHouseholdUseCase = UpdateHouseholdUseCase{}

// UpdateHouseholdUseCase updates mutable household fields.
type UpdateHouseholdUseCase struct {
	repo outbound.HouseholdRepository
}

// NewUpdateHouseholdUseCase constructs UpdateHouseholdUseCase.
func NewUpdateHouseholdUseCase(repo outbound.HouseholdRepository) UpdateHouseholdUseCase {
	return UpdateHouseholdUseCase{repo: repo}
}

// Execute updates the household configuration.
func (u UpdateHouseholdUseCase) Execute(ctx context.Context, input inbound.UpdateHouseholdInput) error {
	h, err := u.repo.FindByID(ctx, input.HouseholdID)
	if err != nil {
		return fmt.Errorf("update household: %w", err)
	}

	if err := h.UpdateConfig(input.Name, input.SettlementMode, input.Currency); err != nil {
		return fmt.Errorf("update household: %w", err)
	}

	if err := u.repo.Update(ctx, h); err != nil {
		return fmt.Errorf("update household: %w", err)
	}

	slog.InfoContext(ctx, "update household", "household_id", input.HouseholdID)
	return nil
}
