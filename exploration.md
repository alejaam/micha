## Exploration: period-guardrails-frontend-adjustments

### Current State

Periods are created in only **two places** in the backend:
1. **`initialize_period.go`** — creates the FIRST period for a household based on `closingDay` and today's date. The period can (and should) extend into the future — it is the *current* period.
2. **`close_period.go`** — when an owner closes a period in `review` status, it creates the *next* period by adding 1 day to the current period's `end_date` and computing a new `end_date` based on `household.PeriodFrequency` (`biweekly` = +14 days, `monthly` = next month - 1 day).

The **domain entity** (`domain/period/period.go`) validates:
- `start_date <= end_date`
- `end_date - start_date >= 7 days`
- Status is valid

It does **NOT** validate against "today" — and correctly so, because the domain layer has no clock dependency.

The frontend does **NOT** calculate next-period dates. It only:
- Displays the current period name (`AppHeader`, `PeriodSelector`)
- Lists historical closed periods (`PeriodHistory`, `PeriodSelector`)
- Calls `closePeriod` API when the owner clicks "Finalizar" (`PeriodManagementPanel`)

### Affected Areas

- `backend/internal/application/period/close_period.go` — **primary guard location**. Before creating `nextPeriod`, check if `nextStart` is in the future.
- `backend/internal/adapters/http/period_handler.go` — map the new error to HTTP 400 Bad Request.
- `backend/internal/domain/shared/errors.go` — add `ErrFuturePeriod` (optional but cleaner).
- `backend/internal/application/period/close_period_test.go` — add test for rejected close when next period is in the future.
- `frontend/src/components/PeriodManagementPanel.jsx` — friendly Spanish error display + optional pre-emptive disable of close button.

### Approaches

1. **Backend-only guard (recommended)**
   - In `ClosePeriodUseCase.Execute`, after calculating `nextStart`, normalize both `nextStart` and `now` to date-only (`00:00:00`) and reject if `nextStart > today`.
   - Return a business-level error: `close period: next period would start in the future`.
   - Handler maps it to 400 Bad Request.
   - Frontend shows the error as-is or with a friendly wrapper.
   - Pros: Single source of truth, works for all clients (web, mobile, API), aligns with Clean Architecture (business rule in use case).
   - Cons: User only discovers the restriction when they try to close.
   - Effort: Low

2. **Backend guard + frontend pre-validation**
   - Same as #1, but frontend also computes whether `nextStart > today` and disables the "Finalizar" button with a tooltip like "El cierre estará disponible a partir del {date}".
   - Pros: Better UX, prevents frustration.
   - Cons: Slightly more code, frontend must duplicate the `nextStart` logic (or derive it from `period.end_date + 1 day`).
   - Effort: Low-Medium

3. **Domain-level clock validation**
   - Inject `now` into `period.New` and reject future start dates at the entity level.
   - Pros: Enforced everywhere.
   - Cons: Violates Clean Architecture — domain shouldn't know about "today" for this rule. A future period is a valid entity; the restriction is a *business workflow* rule (only during close), not an entity invariant. Also breaks `initialize_period.go` which legitimately creates a period ending in the future.
   - Effort: Medium — and architecturally wrong.

### Recommendation

**Approach #1 (backend-only) as MVP, with a touch of #2 for UX.**

- Add the guard in `ClosePeriodUseCase` (use case layer) because this is a **business workflow rule**, not a domain invariant.
- Normalize dates to avoid time-component bugs (existing code sets end-of-day times, so `nextStart` ends up at 23:59:59).
- Add `ErrFuturePeriod` to `shared/errors.go` for clean error identity.
- Map it in `period_handler.go` to 400 Bad Request.
- In `PeriodManagementPanel`, catch the error and show a Spanish message like: *"No se puede cerrar el periodo porque el siguiente comenzaría en el futuro. El cierre estará disponible a partir del {fecha}."*
- Optionally disable the close button when `period.end_date >= today` (derived client-side from the current period's `end_date`).

### Edge Cases

| Scenario | Behavior |
|----------|----------|
| **Today is the last day of the period** | `nextStart = tomorrow > today` → close is rejected. Period stays in `review`. User must wait until tomorrow. This is the intended behavior per the rule. |
| **Switching frequency mid-way** | The next period's duration uses the household's **current** `PeriodFrequency`. The guard still applies: if `nextStart > today`, reject. No special handling needed. |
| **Exact boundary day** | If `today == nextStart` (e.g., period ended yesterday, today is the first day of the next period), close is **allowed** because the next period does not start *beyond* today. |
| **Time-of-day precision** | Must normalize both dates to `00:00:00` before comparing. Otherwise `nextStart` (23:59:59) would be considered "after" noon on the same day. |

### Risks

- **Existing time-component bug in `close_period.go`**: `nextStart = p.EndDate().Add(24 * time.Hour)` where `EndDate` is `23:59:59` results in `nextStart = 23:59:59` of the next day. The new guard MUST normalize dates or it will falsely reject closes on the boundary day.
- **Natural closing in `handleGetCurrent`**: If a period auto-transitions to `review` because its end date passed, but `nextStart > today` (e.g., period ended yesterday and today is still before nextStart? No, nextStart = end+1, so if end passed, nextStart <= today). This is fine.
- **UX confusion**: Users may not understand why they can't close a period that is already in review. The error message must be crystal clear.

### Scope Estimate

| File | Lines changed |
|------|---------------|
| `backend/internal/application/period/close_period.go` | ~8 |
| `backend/internal/domain/shared/errors.go` | ~1 |
| `backend/internal/adapters/http/period_handler.go` | ~4 |
| `backend/internal/application/period/close_period_test.go` | ~35 |
| `frontend/src/components/PeriodManagementPanel.jsx` | ~12 |
| **Total** | **~60 lines** |

Well under the 400-line review budget. No chained PRs needed.

### Ready for Proposal

**Yes.** The scope is small, the rule is unambiguous, and the architecture is clear. The orchestrator can proceed to `sdd-propose` or directly to `sdd-spec` since the approach is straightforward.
