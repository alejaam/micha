package inbound

import "context"

type RegisterExpenseInput struct {
	HouseholdID string
	AmountCents int64
	Description string
}

type RegisterExpenseOutput struct {
	ExpenseID string
}

type RegisterExpenseUseCase interface {
	Execute(ctx context.Context, input RegisterExpenseInput) (RegisterExpenseOutput, error)
}
