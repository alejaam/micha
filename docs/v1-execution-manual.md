# Micha v1.0 - AI Agent Execution Guide

## 1) Executable Roadmap by Phase

### Phase 1 - Stabilization
- Secure runtime configuration and API guardrails.
- Enforce critical business invariants at handler + use case boundaries.
- Build test foundation (backend integration + frontend test harness).
- Fix high-risk data integrity gaps (FKs, constraints, indexes).

### Phase 2 - v1.0 Core Features
- Formalize household role model (owner/admin/member) and enforce it.
- Implement period lifecycle controls for settlement (open/review/closed).
- Improve UX reliability in critical flows (create/edit/delete expense, settlement visibility).

### Phase 3 - Hardening and Scale
- Add observability and operational safeguards.
- Optimize settlement path and repository query performance.
- Add audit trail for sensitive changes.
- Prepare release checklist and production gate.

## 2) Atomic Sprints (Step-by-Step)

### Sprint A1 - Security Baseline
- Technical Objective:
  - Modify `backend/internal/infrastructure/config/config.go` and add tests in `backend/internal/infrastructure/config/config_test.go`.
  - Enforce JWT secret length and production-safe CORS policy.
- Applicable Business Logic:
  - BR-04 Membership scope (indirectly via safer defaults).
  - BR-10 API stability and secure operation baseline.
- Suggested Execution Prompt:
  - "Implement configuration hardening in Go for `config.Load()`: validate `JWT_SECRET >= 32`, require `ALLOWED_ORIGINS` in production, reject wildcard `*` in production, and add table-driven tests using `testify/require` in `config_test.go`. Do not change layer architecture."

### Sprint A2 - Monetary Guardrails
- Technical Objective:
  - Modify `backend/internal/adapters/http/expense_handler.go` and `backend/internal/application/expense/register_expense.go`.
  - Reject `amount_cents <= 0` on create/update before persistence.
- Applicable Business Logic:
  - BR-02 Monetary integrity.
  - BR-03 Positive amounts.
- Suggested Execution Prompt:
  - "Strengthen monetary validations for expenses: in create and patch handlers return `400 INVALID_MONEY` when `amount_cents <= 0`; add HTTP contract tests in `expense_handler_test.go` proving the use case is not called for invalid input."

### Sprint A3 - Repository Integration Tests
- Technical Objective:
  - Create integration tests for `backend/internal/adapters/postgres/expense_repository.go`, `installment_repository.go`, `member_repository.go`.
  - Add test DB bootstrap fixtures and migration setup helper.
- Applicable Business Logic:
  - BR-06 Referential integrity.
  - BR-08 Deterministic settlement period.
- Suggested Execution Prompt:
  - "Create a Postgres integration test suite (Go) for expense/installment/member repositories. Use a temporary DB, run migrations, validate inserts/queries/household-period filters and error paths. Use table-driven tests and `require` for critical assertions."

### Sprint A4 - Frontend Test Foundation
- Technical Objective:
  - Update `frontend/package.json`, create `frontend/vitest.config.*`, add first tests for `src/pages/DashboardPage.jsx` critical flows.
- Applicable Business Logic:
  - BR-04 Membership scope (UI behavior + protected flows).
  - BR-10 API contract stability (error handling).
- Suggested Execution Prompt:
  - "Set up Vitest + Testing Library in the React frontend and add initial Dashboard tests: render, API error state, and unauthorized action blocking. Keep existing components and avoid UX redesign."

### Sprint B1 - Role Authorization Model
- Technical Objective:
  - Modify domain/application/HTTP for role-aware checks in settlement and split config endpoints.
  - Files likely: `backend/internal/domain/member/*`, `backend/internal/application/household/*`, `backend/internal/adapters/http/*handler.go`.
- Applicable Business Logic:
  - BR-05 Privileged actions.
  - BR-04 Membership scope.
- Suggested Execution Prompt:
  - "Implement owner/admin/member role-based authorization for privileged actions (settlement close, split-config update). Add reusable domain errors and tests across handlers/use cases."

### Sprint B2 - Period Lifecycle
- Technical Objective:
  - Add migrations for period persistence and repository/use case support.
  - Create or extend `backend/internal/domain/period/*` and settlement orchestration.
- Applicable Business Logic:
  - BR-08 Deterministic settlement period.
  - BR-01 Net balances only.
- Suggested Execution Prompt:
  - "Implement period lifecycle (open/review/closed) with SQL persistence, state constraints, and settlement integration. Add migration, ports, repository, and tests for valid/invalid transitions."

### Sprint B3 - Dashboard Decomposition
- Technical Objective:
  - Refactor `frontend/src/pages/DashboardPage.jsx` into smaller modules/hooks.
  - Create `frontend/src/hooks/useDashboardState.js` and split UI sections.
- Applicable Business Logic:
  - BR-10 Contract stability (prevent regression by modularity + tests).
- Suggested Execution Prompt:
  - "Refactor DashboardPage to reduce coupling: extract state/actions into a custom hook and split UI sections. Keep current behavior and add regression tests for create/edit/delete expense flows."

### Sprint C1 - Observability and Security Middleware
- Technical Objective:
  - Add middleware for security headers and rate limiting in `backend/internal/adapters/http/`.
  - Wire in `backend/internal/adapters/http/server.go`.
- Applicable Business Logic:
  - BR-09 Auditability.
  - BR-10 API stability.
- Suggested Execution Prompt:
  - "Add HTTP middleware for rate limiting and security headers (HSTS, nosniff, frame deny, baseline CSP), wire it into the existing chain, and cover with middleware tests."

### Sprint C2 - Audit Trail
- Technical Objective:
  - Add migration(s) for `audit_log` and adapter writes for sensitive operations.
  - Target flows: split-config changes, member salary updates, expense delete/patch.
- Applicable Business Logic:
  - BR-07 Soft deletion policy.
  - BR-09 Auditability.
- Suggested Execution Prompt:
  - "Implement sensitive-change auditing: create `audit_log` table and persist before/after plus actor for key financial operations. Add integration tests to guarantee full traceability."

### Sprint C3 - Release Gate
- Technical Objective:
  - Create release checklist doc and CI quality gates (tests + static + security).
  - Files: `docs/release-checklist-v1.md`, CI config file.
- Applicable Business Logic:
  - BR-10 API stability.
- Suggested Execution Prompt:
  - "Create the v1.0 release gate: technical checklist, validation commands, and minimum pipeline (backend/frontend tests, lint, security checks). It must fail if Definition of Done is not met."

## 3) Human Onboarding Guide (Executive Summary)

Start with `.ai-context.md`, then read this manual (`docs/v1-execution-manual.md`). The first file defines technical identity, business invariants, and Definition of Done; the second file turns those constraints into an operational sprint sequence with exact prompts. If an agent proposes changes outside those rules, treat it as architecture or product drift.

The three most dangerous pain points today are incomplete operational security (production configuration, rate limiting, headers), insufficient testing in critical areas (no frontend suite and no DB integration tests for key repositories), and high orchestration complexity in the frontend dashboard that increases regression risk. On the backend, settlement and period/installment repository paths are high-risk because of financial impact.

To validate that an AI agent is working correctly, require evidence at three levels: automated tests for affected packages, exact references to modified files, and explicit mapping to business invariants (BR-xx). If a change does not include relevant tests, violates layers (domain/application/ports/adapters), or breaks API/error code contracts, it should not be accepted as done.
