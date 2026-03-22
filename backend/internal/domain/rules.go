// Package domain contains domain rules that govern Micha's business logic.
// These rules are critical invariants that must be enforced across the application.
package domain

// Critical Business Rules
//
// These rules define the core invariants of Micha's financial model.
// They must be enforced by the domain layer and respected by all use cases.
//
// 1. NON-RETROACTIVE CONTRIBUTION PERCENTAGE
//    The % of contribution of a Member never affects periods anterior to its validFrom date.
//    When a member's contribution % is updated, the new percentage only applies to:
//      - The current open period (if validFrom <= period.startDate)
//      - All future periods
//    Past closed periods retain the contribution % that was active during their lifecycle.
//
// 2. INSTALLMENT PERIOD ASSIGNMENT
//    An Installment lives in the period when it is charged, not when the purchase was made.
//    Example: A purchase made in January with 12 monthly installments will have:
//      - Installment 1 in January
//      - Installment 2 in February
//      - ...
//      - Installment 12 in December
//    Each installment is a separate expense record tied to its respective period.
//
// 3. BALANCE IS DERIVED, NEVER STORED
//    The Balance is always calculated/derived from expenses and contributions.
//    It is never stored as an editable field in the database.
//    Balance = sum(member's expenses) - (sum(household expenses) * member's contribution %)
//    The balance is recalculated in real-time whenever:
//      - A new expense is registered
//      - An expense is updated or deleted
//      - A member's contribution % changes (only affects current/future periods)
//
// 4. PERIOD CLOSURE REQUIRES CONSENSUS
//    The closure of a Period requires PeriodApproval from all Members.
//    The flow is:
//      - Any member can initiate closure → period.status = "review"
//      - During "review", no new expenses can be added
//      - Each member must approve or object (with optional comment)
//      - If all members approve → period.status = "closed"
//      - If any member objects → household owner can force closure or reopen for edits
//      - When closed, a new period is automatically generated
//      - Fixed expenses and active installments are replicated to the new period
//
// 5. ONE OPEN PERIOD PER HOUSEHOLD
//    Only one Period with status="open" can exist per Household at any time.
//    This ensures:
//      - Clear temporal boundaries for expense tracking
//      - Deterministic balance calculations
//      - Simplified consensus workflow
//
// 6. EXPENSE REGISTRATION DURING REVIEW
//    During period status="review", no new expenses can be added.
//    Members must either:
//      - Approve the current state and close the period
//      - Object and request the owner to reopen for edits
//    This prevents race conditions during the approval process.
//
// 7. FIXED EXPENSES AND INSTALLMENTS ARE REPLICATED
//    When a period is closed:
//      - All expenses with type="fixed" are copied to the new period
//      - All active installments (currentInstallment < totalInstallments) are copied
//        with currentInstallment incremented by 1
//      - Variable expenses are NOT copied (they are period-specific)
//
// 8. MEMBER REGISTERS OWN EXPENSES
//    Each member registers their own variable expenses.
//    Fixed expenses are typically registered by the household owner or a designated admin.
//    This promotes:
//      - Transparency: everyone sees who paid what
//      - Autonomy: members control their own records
//      - Accountability: no disputes about "who entered what"
//
// 9. MICHA MANAGES SHARED FINANCES ONLY
//    Micha does NOT manage personal finances.
//    It only tracks what is shared within the Household:
//      - Shared expenses (rent, utilities, groceries, etc.)
//      - Individual contributions to the shared pool
//      - The balance between members (who owes whom)
//    Personal expenses, savings, investments, etc. are out of scope.
