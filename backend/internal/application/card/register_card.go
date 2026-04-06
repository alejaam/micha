package cardapp

import (
	"context"
	"fmt"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/card"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

// RegisterCardUseCase creates a new credit card.
type RegisterCardUseCase struct {
	repo        outbound.CardRepository
	idGenerator appshared.IDGenerator
	now         func() time.Time
}

// NewRegisterCardUseCase constructs RegisterCardUseCase.
func NewRegisterCardUseCase(repo outbound.CardRepository, idGenerator appshared.IDGenerator) RegisterCardUseCase {
	return RegisterCardUseCase{
		repo:        repo,
		idGenerator: idGenerator,
		now:         time.Now,
	}
}

// Execute creates a card and stores it.
func (u RegisterCardUseCase) Execute(ctx context.Context, input inbound.RegisterCardInput) (inbound.RegisterCardOutput, error) {
	c, err := card.New(
		card.ID(u.idGenerator.NewID()),
		input.HouseholdID,
		input.BankName,
		input.CardName,
		input.CutoffDay,
		u.now(),
	)
	if err != nil {
		return inbound.RegisterCardOutput{}, fmt.Errorf("register card: %w", err)
	}

	if err := u.repo.Save(ctx, c); err != nil {
		return inbound.RegisterCardOutput{}, fmt.Errorf("register card: %w", err)
	}

	return inbound.RegisterCardOutput{CardID: string(c.ID())}, nil
}

var _ inbound.RegisterCardUseCase = RegisterCardUseCase{}
