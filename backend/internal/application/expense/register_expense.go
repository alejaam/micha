package expenseapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"micha/backend/internal/domain/expense"
)

type IDGenerator interface {
	NewID() string
}

type RegisterExpenseUseCase struct {
	repo        ExpenseRepository
	idGenerator IDGenerator
	now         func() time.Time
}

func NewRegisterExpenseUseCase(repo ExpenseRepository, idGenerator IDGenerator) RegisterExpenseUseCase {
	return RegisterExpenseUseCase{
		repo:        repo,
		idGenerator: idGenerator,
		now:         time.Now,
	}
}

func (u RegisterExpenseUseCase) Execute(ctx context.Context, input RegisterExpenseInput) (RegisterExpenseOutput, error) {
	e, err := expense.New(
		expense.ID(u.idGenerator.NewID()),
		input.HouseholdID,
		input.AmountCents,
		input.Description,
		u.now(),
	)
	if err != nil {
		return RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	if err := u.repo.Save(ctx, e); err != nil {
		return RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	slog.InfoContext(ctx, "register expense", "expense_id", string(e.ID()))
	return RegisterExpenseOutput{ExpenseID: string(e.ID())}, nil
}
