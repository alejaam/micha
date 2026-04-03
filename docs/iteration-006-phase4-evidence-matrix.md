# Iteration 006 - Phase 4/5 Evidence Matrix

Date: 2026-04-03
Scope: Closure evidence for recurring expenses and MSI behavior

## Criteria Matrix

| Criterion | Status | Evidence | Notes |
|----------|--------|----------|-------|
| Recurring templates persist in DB | PASS | `backend/internal/adapters/postgres/recurring_expense_repository.go` | CRUD operations implemented |
| Recurring generation use case exists | PASS | `backend/internal/application/recurringexpense/generate_recurring_expenses.go` | Includes generation loop and progress counters |
| Idempotency guard for recurring generation | PASS | `backend/internal/application/recurringexpense/generate_recurring_expenses.go` (`next_generation_date` advancement) | Re-run should not duplicate same period once date advances |
| Recurring generation HTTP endpoint wired | PASS | `backend/internal/adapters/http/recurring_expense_handler.go`, `backend/internal/adapters/http/server.go` | Endpoint available for UI trigger |
| Expense typing supports `fixed`, `variable`, `msi` | PASS | `backend/internal/domain/expense/expense.go`, `backend/internal/adapters/http/expense_handler.go` | Validation and serialization present |
| Installments persisted for MSI expenses | PASS | `backend/internal/application/expense/register_expense.go`, `backend/internal/adapters/postgres/installment_repository.go` | `SaveAll` called for MSI |
| Settlement reads installments by period | PASS | `backend/internal/application/settlement/calculate_settlement.go` | Calls `ListByHouseholdAndPeriod` |
| Frontend trigger for recurring generation | PASS | `frontend/src/pages/DashboardPage.jsx` (`handleGenerateRecurring`) | User can generate fixed expenses from planning section |
| Frontend visibility for MSI/fixed/variable | PASS | `frontend/src/components/ExpenseItem.jsx` | Type badges rendered in recent/activity flows |
| UI filters by type in planning panels | PARTIAL | `frontend/src/components/FixedExpensesPanel.jsx`, `frontend/src/components/CardExpensesPanel.jsx` | Covered in specialized panels; explicit global type filter control is still open |

Status values: `PASS`, `PARTIAL`, `FAIL`, `PENDING`.

## Open Verification Actions
- Verify idempotency with an explicit repeat-run API test scenario and record response snapshots.
- Validate MSI edge cases across month boundaries with deterministic fixture data.
- Decide whether "filters by type in UI" criterion is satisfied by current sectioned panels or needs a dedicated filter control.
