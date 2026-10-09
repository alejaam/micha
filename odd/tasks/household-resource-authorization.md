# Household Resource Authorization

## Goal

Prevent authenticated users from reading or mutating expenses and recurring expenses that belong to households where they are not active members.

## Confirmed test seams

- Application use-case interfaces: authorization is enforced after resolving the target resource and before returning or mutating it.
- HTTP handlers: the authenticated user ID is propagated to application inputs for every protected operation.

## Tasks

- [x] **1. Protect expense operations** — implemented, independently verified, and committed.
  - Added application tests for cross-household get, list, patch, delete, and period-filter access.
  - Added focused HTTP coverage proving the authenticated user ID reaches expense operations.
  - Enforced active membership against the resolved expense household and verified period ownership before filtered listing.
  - Verification: `go test ./internal/application/expense`, `go test ./internal/adapters/http`, `go test ./...`, and `go test -race ./...` passed from `backend/`.
  - Runtime harness: N/A; authorization is covered at application and HTTP seams, and no live PostgreSQL harness is configured.
  - Independent verification: passed with no blocking findings.
  - Evidence: commit `8fe3f49` (`fix(authz): isolate expense operations by household`).

- [ ] **2. Protect recurring-expense operations**
  - Add failing application tests for cross-household create, get, list, update, delete, and generation access as applicable.
  - Add focused HTTP tests proving authenticated identity propagation.
  - Implement the smallest contract and authorization changes needed to pass.
  - Verification: focused Go tests plus runtime boundary where available.
  - Evidence: pending.

- [ ] **3. Verify isolation and architecture**
  - Run the backend test suite and race detector.
  - Validate dependency direction against `docs/architecture-checklist.md`.
  - Confirm no protected resource path relies only on client-supplied household identity.
  - Evidence: pending.

## Constraints

- Do not implement the deferred MSI, split-config, period-closing, or generic HTTP-validation backlog.
- Preserve Clean Architecture dependency direction.
- Do not alter unrelated existing working-tree changes.
- Do not commit until explicit user authorization is provided.

## Work-unit plan

1. `fix(authz): isolate expense operations by household membership`
2. `fix(authz): isolate recurring expenses by household membership`
3. Verification evidence update, folded into the relevant behavior commit or a documentation-only follow-up if needed.
