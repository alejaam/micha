package expenseapp

import (
	"context"
	"time"

	"micha/backend/internal/domain/expense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type IDGenerator interface {
	NewID() string
}

type RegisterExpenseUseCase struct {
	repo        outbound.ExpenseRepository
	idGenerator IDGenerator
	now         func() time.Time
}

func NewRegisterExpenseUseCase(repo outbound.ExpenseRepository, idGenerator IDGenerator) RegisterExpenseUseCase {
	return RegisterExpenseUseCase{
		repo:        repo,
		idGenerator: idGenerator,
		now:         time.Now,
	}
}

func (u RegisterExpenseUseCase) Execute(ctx context.Context, input inbound.RegisterExpenseInput) (inbound.RegisterExpenseOutput, error) {
	e, err := expense.New(
		expense.ID(u.idGenerator.NewID()),
		input.HouseholdID,
		input.AmountCents,
		input.Description,
		u.now(),
	)
	if err != nil {
		return inbound.RegisterExpenseOutput{}, err
	}

	if err := u.repo.Save(ctx, e); err != nil {
		return inbound.RegisterExpenseOutput{}, err
	}

	return inbound.RegisterExpenseOutput{ExpenseID: string(e.ID())}, nil
}
