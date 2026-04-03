# Iteration 006 - Risk Register

Date: 2026-04-03
Scope: `Phase 4 closure + Phase 6 kickoff + hardening`

## Risk Matrix

| ID | Risk | Impact | Probability | Owner | Mitigation | Trigger |
|----|------|--------|-------------|-------|------------|---------|
| R1 | Documentation drift reappears after code changes | High | High | Copilot + User | Update tracker/roadmap in the same iteration as code changes | Any merged feature without docs touch |
| R2 | Phase 4 marked closed without idempotency evidence | High | Medium | Copilot + User | Require explicit recurring-generation verification checklist before closure | Missing pass/fail evidence matrix |
| R3 | Phase 6 scope expands too early (analytics overreach) | Medium | High | Copilot + User | Keep one KPI vertical slice only; defer trend dashboards | More than one new analytics endpoint in the same sprint |
| R4 | Frontend regressions due to low automated coverage | High | Medium | Copilot + User | Maintain smoke checklist and define test harness backlog as explicit debt | Reopened UI bugs in critical expense/settlement flows |
| R5 | Hidden data integrity edge cases in period-level calculations | High | Medium | Copilot + User | Prioritize settlement and installment edge-case checks before phase closure | Inconsistent settlement values between periods |

## Active Mitigation Tasks
- M1. Keep backend quality gates green after each functional change.
- M2. Publish pass/fail evidence matrix for Phase 4/5 closure.
- M3. Keep Phase 6 to one KPI endpoint plus one dashboard card.
- M4. Capture unresolved frontend test debt as explicit backlog items.

## Review Cadence
- Daily: update risk status (`Open`, `Watching`, `Mitigated`, `Closed`).
- Weekly: review impact/probability and adjust mitigation plan.
