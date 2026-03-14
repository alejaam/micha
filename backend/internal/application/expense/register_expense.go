package expenseapp

import (
	"context"
	"fmt"
	"log/slog"
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
	now := u.now()
	e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
		ID:             expense.ID(u.idGenerator.NewID()),
		HouseholdID:    input.HouseholdID,
		PaidByMemberID: input.PaidByMemberID,
		AmountCents:    input.AmountCents,
		Description:    input.Description,
		IsShared:       input.IsShared,
		Currency:       input.Currency,
		PaymentMethod:  expense.PaymentMethod(input.PaymentMethod),
		ExpenseType:    expense.ExpenseType(input.ExpenseType),
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	if err := u.repo.Save(ctx, e); err != nil {
		return inbound.RegisterExpenseOutput{}, fmt.Errorf("register expense: %w", err)
	}

	slog.InfoContext(ctx, "register expense", "expense_id", string(e.ID()))
	return inbound.RegisterExpenseOutput{ExpenseID: string(e.ID())}, nil
}
