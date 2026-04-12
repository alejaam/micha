package recurringexpenseapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	appshared "micha/backend/internal/application/shared"
	"micha/backend/internal/domain/expense"
	"micha/backend/internal/domain/recurringexpense"
	"micha/backend/internal/ports/inbound"
	"micha/backend/internal/ports/outbound"
)

type GenerateRecurringExpensesUseCase struct {
	recurringExpenseRepo outbound.RecurringExpenseRepository
	expenseRepo          outbound.ExpenseRepository
	idGenerator          appshared.IDGenerator
	now                  func() time.Time
}

func NewGenerateRecurringExpensesUseCase(
	recurringExpenseRepo outbound.RecurringExpenseRepository,
	expenseRepo outbound.ExpenseRepository,
	idGenerator appshared.IDGenerator,
) GenerateRecurringExpensesUseCase {
	return GenerateRecurringExpensesUseCase{
		recurringExpenseRepo: recurringExpenseRepo,
		expenseRepo:          expenseRepo,
		idGenerator:          idGenerator,
		now:                  time.Now,
	}
}

func (u GenerateRecurringExpensesUseCase) Execute(ctx context.Context, cmd inbound.GenerateRecurringExpensesCommand) (inbound.GenerateRecurringExpensesOutput, error) {
	asOfDate := cmd.AsOfDate
	if asOfDate.IsZero() {
		asOfDate = u.now()
	}

	// Find all recurring expenses that are due for generation
	recurringExpenses, err := u.recurringExpenseRepo.ListDueForGeneration(ctx, asOfDate)
	if err != nil {
		return inbound.GenerateRecurringExpensesOutput{}, fmt.Errorf("generate recurring expenses: %w", err)
	}

	// Filter by household if specified
	if cmd.HouseholdID != "" {
		filtered := make([]recurringexpense.RecurringExpense, 0)
		for _, re := range recurringExpenses {
			if re.HouseholdID() == cmd.HouseholdID {
				filtered = append(filtered, re)
			}
		}
		recurringExpenses = filtered
	}

	var generatedIDs []string
	for _, re := range recurringExpenses {
		if re.IsAgnostic() {
			continue
		}

		// Double-check that this recurring expense should generate an expense
		if !re.ShouldGenerateExpense(asOfDate) {
			continue
		}

		// Create the expense from the recurring template
		e, err := expense.NewFromAttributes(expense.ExpenseAttributes{
			ID:             expense.ID(u.idGenerator.NewID()),
			HouseholdID:    re.HouseholdID(),
			PaidByMemberID: re.PaidByMemberID(),
			AmountCents:    re.AmountCents(),
			Description:    fmt.Sprintf("%s (auto)", re.Description()),
			IsShared:       true,
			Currency:       "MXN",
			PaymentMethod:  expense.PaymentMethodCash,
			ExpenseType:    re.ExpenseType(),
			CardName:       "",
			CategoryID:     re.CategoryID(),
			CreatedAt:      asOfDate,
			UpdatedAt:      asOfDate,
		})
		if err != nil {
			slog.ErrorContext(ctx, "generate recurring expenses: failed to create expense",
				"recurring_expense_id", string(re.ID()),
				"error", err,
			)
			continue
		}

		// Save the generated expense
		if err := u.expenseRepo.Save(ctx, e); err != nil {
			slog.ErrorContext(ctx, "generate recurring expenses: failed to save expense",
				"recurring_expense_id", string(re.ID()),
				"expense_id", string(e.ID()),
				"error", err,
			)
			continue
		}

		generatedIDs = append(generatedIDs, string(e.ID()))

		// Advance the next generation date for idempotency
		re.AdvanceNextGenerationDate()
		if err := u.recurringExpenseRepo.Update(ctx, re); err != nil {
			slog.ErrorContext(ctx, "generate recurring expenses: failed to update recurring expense",
				"recurring_expense_id", string(re.ID()),
				"error", err,
			)
			// Continue even if update fails - the expense was already created
		}

		slog.InfoContext(ctx, "generated expense from recurring template",
			"recurring_expense_id", string(re.ID()),
			"expense_id", string(e.ID()),
			"next_generation_date", re.NextGenerationDate(),
		)
	}

	return inbound.GenerateRecurringExpensesOutput{
		GeneratedCount: len(generatedIDs),
		ExpenseIDs:     generatedIDs,
	}, nil
}

var _ inbound.GenerateRecurringExpensesUseCase = GenerateRecurringExpensesUseCase{}
