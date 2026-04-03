# Iteration 007 - Onboarding, Settings, and Recurring UX Plan

Date: 2026-04-03  
Branch: `feature/improving-ux`

## Goal
Ship an incremental UX and behavior fix set for onboarding, household settings visibility, member invite flow, and recurring expense automation, with explicit state tracking in Engram.

## Verified Findings (from current code)

1. Cards onboarding is effectively mandatory.
- `frontend/src/pages/OnboardingHouseholdPage.jsx` always navigates to `/onboarding/cards`.
- `frontend/src/pages/OnboardingCardsPage.jsx` shows skip messaging but disables final continue when there are no cards.

2. Duplicate CTA text exists in cards onboarding.
- `frontend/src/pages/OnboardingCardsPage.jsx` renders two buttons with "Continue to fixed expenses" in the same action row.

3. Dashboard currently exposes payment-weight information in the main view.
- `frontend/src/components/MembersPanel.jsx` renders salary and member "weight %".
- `frontend/src/pages/DashboardPage.jsx` includes members and incomes in overview by default.

4. Household split configuration API exists but no dedicated settings UI route/page exists in frontend.
- Backend supports `GET /v1/households/{household_id}`, `PUT /v1/households/{household_id}`, and `PUT /v1/households/{household_id}/split-config`.
- Frontend `src/api.js` currently does not expose `getHousehold`, `updateHousehold`, or `updateSplitConfig` helpers.

5. Member invite capability exists in backend and in a generic frontend page, but not as a first-class post-onboarding owner workflow.
- Route `POST /v1/households/{household_id}/members` already exists.
- Frontend has `OnboardingMemberPage` and route `/members/new`.

6. Default payment method is already cash in both frontend and backend.
- `frontend/src/api.js` createExpense defaults `paymentMethod = 'cash'`.
- Domain and recurring generation default to cash behavior.

7. Recurring/fixed generation exists but is currently manual from dashboard.
- `frontend/src/pages/DashboardPage.jsx` exposes "Generar gastos fijos" button.
- Desired behavior is automatic monthly generation unless disabled/changed in edit settings.

## Scope Breakdown by Iteration

## Iteration 007.1 - Onboarding flow cleanup (low risk, high UX value)

### Scope
- Make cards step optional.
- Remove duplicated CTA in cards onboarding.
- Keep fixed-expenses step optional and reachable whether cards were added or not.

### Changes
- Frontend only.
- Add explicit `Skip cards for now` action in `OnboardingCardsPage`.
- Ensure one single primary continue CTA.
- Confirm forward flow: household -> cards (optional) -> fixed expenses (optional) -> dashboard.

### Acceptance Criteria
- User can complete onboarding without creating cards.
- No duplicated "Continue to fixed expenses" buttons.
- Navigation remains deterministic (no dead-ends).

### Risk
- Minor navigation regressions.

## Iteration 007.2 - Dashboard information architecture (main-page simplification)

### Scope
- Remove settings-oriented payment weight/configuration from the main page.
- Show split scheme at the top in a concise way.

### Changes
- Frontend only.
- Add a compact "Split Scheme" summary block near top of overview.
- Reduce or move salary/percentage detail from default dashboard panels.

### Acceptance Criteria
- Main page no longer emphasizes payment percentage/configuration details.
- Split mode summary is visible above fold.

### Risk
- Users may temporarily lose access to detailed settings if settings page is not shipped in the same iteration.

## Iteration 007.3 - New Household Settings section/page

### Scope
- Create a dedicated settings area to view/edit household setup values created at onboarding.
- Include settlement mode, currency, and split percentages.

### Changes
- Frontend + API wiring (existing backend endpoints).
- Add frontend API methods:
  - `getHousehold({ householdId })`
  - `updateHousehold({ householdId, name, settlementMode, currency })`
  - `updateSplitConfig({ householdId, splits })`
- Add route/page, e.g. `/household/settings`.
- Add dashboard link/button to open settings.

### Acceptance Criteria
- User can open settings and review initial household configuration.
- User can update settlement mode/currency.
- User can update split percentages and backend persists correctly.

### Risk
- Validation edge cases (split sum must equal 100).

## Iteration 007.4 - Owner member management workflow

### Scope
- Ensure household creator/owner can add/invite members after onboarding from a clear path.

### Changes
- Frontend UX and routing.
- Promote `/members/new` action into obvious dashboard CTA and/or settings section.
- Keep existing backend member creation contract.

### Acceptance Criteria
- Household owner can reliably add new members after onboarding without hidden routes.
- New members appear in list after creation.

### Risk
- Role UX confusion if owner/admin semantics are not explicit in copy.

## Iteration 007.5 - Automatic monthly recurring generation

### Scope
- Recurring fixed expenses should auto-generate month by month by default.
- Manual generation remains as fallback.
- Ability to disable/adjust auto behavior from settings/edit flow.

### Changes
- Backend + frontend.
- Introduce a recurring automation policy (default enabled).
- Trigger idempotent generation automatically on selected household/month entry.
- Keep "Generate now" button as explicit retry/manual control.

### Acceptance Criteria
- Opening dashboard in a new month auto-generates due recurring expenses once.
- Re-entering dashboard does not create duplicates.
- User can disable or adjust automation behavior from settings.

### Risk
- Duplicate generation if idempotency boundaries are incomplete.

## Technical Notes

- Keep feature order per architecture for backend additions:
  - domain -> application -> ports -> adapters -> wiring
- Reuse existing split config and recurring generation use cases where possible.
- Avoid introducing hidden business logic in frontend-only guards.

## Engram Execution Protocol for This Plan

For each iteration close-out, persist one observation with:
- title: `Complete iteration 007.X <short-topic>`
- type: `decision` or `bugfix` or `discovery`
- topic_key: `roadmap/iteration-007`
- content sections:
  - **What**
  - **Why**
  - **Where**
  - **Learned**

Suggested progress checkpoints:
- `roadmap/iteration-007/backlog-locked`
- `roadmap/iteration-007/in-progress`
- `roadmap/iteration-007/validation`
- `roadmap/iteration-007/released`

## Recommended Implementation Order

1. Iteration 007.1 (onboarding cleanup)
2. Iteration 007.2 (dashboard simplification + split summary)
3. Iteration 007.3 (dedicated settings)
4. Iteration 007.4 (owner member invites from clear path)
5. Iteration 007.5 (auto recurring generation)

This order minimizes UX confusion while preparing the settings surface before introducing recurring automation controls.

## Implementation Status Snapshot (2026-04-03)

- `007.1` Onboarding cards optional flow: `DONE`.
- `007.2` Dashboard simplification + split summary: `DONE`.
- `007.3` Household settings route + API wiring: `DONE`.
- `007.4` Owner member workflow from visible paths: `DONE`.
- `007.5` Recurring monthly automation policy: `DONE (frontend-first)`.

Notes:
- `007.5` currently persists automation policy on frontend per household (`localStorage`) and keeps backend generation endpoint unchanged.
- Manual generation remains available as explicit fallback.

## Validation Snapshot (Smoke Checklist - 2026-04-03)

Validation type: `code-path smoke review` (no runtime E2E execution in this pass).

| Scenario | Expected | Status | Evidence |
|----------|----------|--------|----------|
| Owner can find member invite flow from dashboard | Clear CTA to member creation route | PASS | `frontend/src/pages/DashboardPage.jsx` (member card + `navigate('/members/new')`) |
| Owner can find member invite flow from settings | Clear CTA in household settings page | PASS | `frontend/src/pages/HouseholdSettingsPage.jsx` (`Member management` section) |
| Member creation route is reachable | `/members/new` route remains wired | PASS | `frontend/src/router.jsx` |
| Recurring automation can be enabled/disabled in settings | Toggle writes per-household preference | PASS | `frontend/src/pages/HouseholdSettingsPage.jsx`, `frontend/src/utils.js` |
| Dashboard auto-generates recurring once per period | Trigger guarded by household+period key | PASS | `frontend/src/pages/DashboardPage.jsx`, `frontend/src/utils.js` |
| Manual recurring generation still works | Manual button calls generation endpoint | PASS | `frontend/src/pages/DashboardPage.jsx` (`handleGenerateRecurring`) |

## Residual Risks

- Frontend smoke validation is based on code paths, not browser-executed E2E scenarios.
- Recurring automation preference is client-scoped (browser storage), not yet centralized in backend household settings.
