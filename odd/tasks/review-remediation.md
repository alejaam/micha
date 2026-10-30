# Feature: Review Remediation (10 Audit Findings)

Remediation of 10 audit findings covering backend migration safety, onboarding fixed expenses integrity, theme flash/toggle UX, and minor component refactoring.

## Specs

- S1: "Baseline migration vs. existing databases (migrations.go:67): the 33 old migrations were replaced by 001_initial_schema.sql. A database that already ran the old files will treat the baseline as unapplied. CREATE TABLE IF NOT EXISTS then skips existing tables, so any schema differences go unapplied without a warning. Any non-idempotent statement would abort startup. You need either a reconcile or baseline-detection step, or a documented "recreate the DB" requirement."
- S2: "Empty checksum is backfilled (migrations.go:65): a row with an empty checksum gets the hash of whatever file is on disk. That skips the tamper check for exactly those rows."
- S3: "Advisory lock on a separate connection (migrations.go:24): the lock is held on one pooled connection, but the migration work runs on other connections. With MaxConns=1, startup would deadlock. If ctx is cancelled, the unlock fails silently and the connection goes back to the pool still holding the lock."
- S4: "Onboarding fixed expenses: partial saves (OnboardingFixedExpensesPage.jsx:196): handleSave makes several create calls with no rollback. If one fails, a retry duplicates the ones that already succeeded. The hasValidSelection guard was also removed."
- S5: "Half-filled custom entry is dropped silently (OnboardingFixedExpensesPage.jsx:132): if the name or amount is invalid, the custom expense is ignored and the success message still shows. If only the custom entry is half-filled, no Save button appears and nothing explains why."
- S6: "Custom expense payload (OnboardingFixedExpensesPage.jsx:237): it is sent with paidByMemberId: ''. If the backend requires a payer UUID, this fails after the catalog items have already saved, which leads to the partial-save problem in item 4."
- S7: "Flash of the wrong theme (ThemeContext.jsx:54): data-theme is only set in a useEffect. A user who saved "light" on a dark-OS device gets a dark flash on every load. The localStorage write inside the state updater is also impure, and StrictMode double-invokes updaters. Fix it with an inline script in index.html that sets the attribute before first paint."
- S8: "Quick toggle (AppHeader.jsx:160): it cycles auto → dark → light. When the OS is dark, the first click looks like a no-op, and the icon doesn't show "auto" mode."
- S9: "autoFocus on every amount input (OnboardingFixedExpensesPage.jsx:497): it steals focus on each toggle and opens the keyboard on mobile."
- S10: "Duplicated form branches (OnboardingHouseholdPage.jsx:131): the onboarding and edit branches repeat the same form. Extract one shared component."

## Tasks

- [ ] T1: Fix backend migrations advisory lock, single connection execution, and baseline reconciliation/checksum rules (Linked S1, S2, S3) [inline]
- [ ] T2: Fix onboarding fixed expenses partial save retry, selection guard, half-filled custom validation, and remove autoFocus (Linked S4, S5, S6, S9) [inline]
- [ ] T3: Fix theme initial flash with index.html inline script, purify updater, and improve quick toggle icon & cycle (Linked S7, S8) [inline]
- [ ] T4: Refactor OnboardingHouseholdPage to extract duplicated form branches into a shared component (Linked S10) [inline]

## Log

- L1: User requested review of 10 findings:
  1. Baseline migration vs. existing databases (migrations.go:67)
  2. Empty checksum is backfilled (migrations.go:65)
  3. Advisory lock on a separate connection (migrations.go:24)
  4. Onboarding fixed expenses: partial saves (OnboardingFixedExpensesPage.jsx:196)
  5. Half-filled custom entry is dropped silently (OnboardingFixedExpensesPage.jsx:132)
  6. Custom expense payload (OnboardingFixedExpensesPage.jsx:237)
  7. Flash of the wrong theme (ThemeContext.jsx:54)
  8. Quick toggle (AppHeader.jsx:160)
  9. autoFocus on every amount input (OnboardingFixedExpensesPage.jsx:497)
  10. Duplicated form branches (OnboardingHouseholdPage.jsx:131)
- L2: User confirmed to implement all 10 fixes in this session.
