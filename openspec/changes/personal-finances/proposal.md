# Proposal: Personal Finances & Credit Cards

## Intent

Support personal (member-scoped) credit cards and individual financial tracking by calculating a member's remaining salary. This ensures personal expenses are separated from household settlements while allowing personal cards to pay for shared expenses.

## Scope

### In Scope
- Change cards from household-scoped to member-scoped (`owner_member_id`).
- Add business rule: Only the card owner can register an expense using their card.
- Add default categories: "Streaming/Services" and "Personal".
- Add business rule: Force `is_shared = false` for expenses with the "Personal" category (or explicitly marked personal).
- Add new capability/endpoint to calculate a member's remaining salary for a given period: `Salary - (Sum of Personal Expenses + Allocated portion of Shared Expenses)`.

### Out of Scope
- Modifying "Fixed" and "Installments (MSI)", which remain as Expense Types (not Categories).
- Frontend/UI implementation (this is backend-focused).

## Capabilities

### New Capabilities
- `personal-cards`: Member-scoped card ownership and usage validation rules.
- `remaining-salary`: Calculation of a member's available funds based on fixed salary minus personal and shared financial obligations.

### Modified Capabilities
- None (no existing specs to update).

## Approach

1. **Domain**: 
   - Update the `Card` entity to require an `OwnerMemberID`.
   - Update the `Expense` entity (or a domain service) to enforce two invariants:
     - If a `CardID` is provided, the `PayerMemberID` must match the `Card.OwnerMemberID`.
     - If the category is "Personal", force `IsShared = false`.
2. **Use Cases**:
   - Update `CreateExpense` to fetch the Card and validate the owner against the user initiating the request.
   - Add a new `CalculateRemainingSalary` use case that aggregates personal expenses and the member's share of shared expenses for a specific period.
3. **Adapters**:
   - Create Postgres migrations to add `owner_member_id` to the `cards` table.
   - Add the new categories to the database (or configuration, depending on how categories are currently managed).
   - Create a new HTTP endpoint for retrieving the remaining salary.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/domain/card` | Modified | Add `OwnerMemberID` to entity. |
| `internal/domain/expense` | Modified | Add logic to force `IsShared=false` for "Personal" category. |
| `internal/application/expense` | Modified | `CreateExpense` use case must load the Card and validate ownership. |
| `internal/application/member` | New | Add `CalculateRemainingSalary` use case. |
| `internal/adapters/postgres` | Modified | Add DB migrations for `cards` and `categories` tables, and implement SQL queries for remaining salary. |
| `internal/adapters/http` | Modified | Add new route and handler for remaining salary. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Data migration issues with existing household cards. | Medium | Write a safe migration that assigns existing cards to a default member or requires manual intervention before applying. |
| Incorrect remaining salary calculation. | High | Implement comprehensive unit tests for the calculation logic across different edge cases (e.g., MSI, different split strategies). |

## Rollback Plan

- Revert the database migrations (drop `owner_member_id` from `cards`).
- Revert application changes (remove ownership validation, new endpoints, and use cases).

## Dependencies

- None.

## Success Criteria

- [ ] A member can only create an expense using their own personal card.
- [ ] Expenses created with the "Personal" category automatically have `is_shared` set to `false`.
- [ ] The remaining salary endpoint correctly returns: `Monthly Salary - (Personal Expenses + Share of Shared Expenses)`.
- [ ] Settlement calculations are unaffected by personal expenses.