package householdapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"micha/backend/internal/domain/household"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// IDGenerator abstracts id generation for testability.
type IDGenerator interface {
	NewID() string
}

// RegisterHouseholdUseCase creates a new household.
type RegisterHouseholdUseCase struct {
	repo        outbound.HouseholdRepository
	idGenerator IDGenerator
	now         func() time.Time
}

// NewRegisterHouseholdUseCase constructs RegisterHouseholdUseCase.
func NewRegisterHouseholdUseCase(repo outbound.HouseholdRepository, idGenerator IDGenerator) RegisterHouseholdUseCase {
	return RegisterHouseholdUseCase{repo: repo, idGenerator: idGenerator, now: time.Now}
}

// Execute creates a household and stores it.
func (u RegisterHouseholdUseCase) Execute(ctx context.Context, input inbound.RegisterHouseholdInput) (inbound.RegisterHouseholdOutput, error) {
	h, err := household.New(
		household.ID(u.idGenerator.NewID()),
		input.Name,
		input.SettlementMode,
		input.Currency,
		u.now(),
	)
	if err != nil {
		return inbound.RegisterHouseholdOutput{}, fmt.Errorf("register household: %w", err)
	}

	if err := u.repo.Save(ctx, h); err != nil {
		return inbound.RegisterHouseholdOutput{}, fmt.Errorf("register household: %w", err)
	}

	slog.InfoContext(ctx, "register household", "household_id", string(h.ID()))
	return inbound.RegisterHouseholdOutput{HouseholdID: string(h.ID())}, nil
}
