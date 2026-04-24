# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

See `_shared/skill-resolver.md` for the full resolution protocol.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| When creating a pull request, opening a PR, or preparing changes for review. | branch-pr | /home/ale/.copilot/skills/branch-pr/SKILL.md |
| When writing Go tests, using teatest, or adding test coverage. | go-testing | /home/ale/.copilot/skills/go-testing/SKILL.md |
| — (description-only trigger) | golang-patterns | /home/ale/.copilot/skills/golang-patterns/SKILL.md |
| — (description-only trigger) | golang-testing | /home/ale/.copilot/skills/golang-testing/SKILL.md |
| When creating a GitHub issue, reporting a bug, or requesting a feature. | issue-creation | /home/ale/.copilot/skills/issue-creation/SKILL.md |
| When user says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen". | judgment-day | /home/ale/.copilot/skills/judgment-day/SKILL.md |
| Use when the user asks for Nothing-style, transparent-hardware, or monochrome minimal interfaces. | nothing-ui-shell | /home/ale/.copilot/skills/nothing-ui-shell/SKILL.md |
| — (description-only trigger) | postgres-patterns | /home/ale/.copilot/skills/postgres-patterns/SKILL.md |
| — (description-only trigger) | reviewer | /home/ale/.copilot/skills/reviewer/SKILL.md |
| When user asks to create a new skill, add agent instructions, or document patterns for AI. | skill-creator | /home/ale/.copilot/skills/skill-creator/SKILL.md |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### branch-pr
- Every PR MUST link an approved issue (`status:approved`).
- PR MUST include exactly one `type:*` label.
- Branch names MUST match `type/description` with lowercase slug.
- Use conventional commit format; keep type aligned with PR label.
- Include issue linkage (`Closes/Fixes/Resolves #N`) in PR body.
- Run required checks before merge; blocked checks are release blockers.

### go-testing
- Prefer table-driven tests with subtests for variant coverage.
- Test success and error paths explicitly; avoid partial assertions.
- For Bubbletea, test model transitions with `Update` and key messages.
- Use `teatest` for interactive end-to-end TUI flows.
- Use golden files for stable rendering output checks.
- Use `t.TempDir()` and deterministic fixtures for filesystem/system tests.

### golang-patterns
- Keep code simple and explicit; prefer clarity over cleverness.
- Accept interfaces and return concrete structs.
- Wrap errors with context and use `errors.Is/As` for matching.
- Propagate `context.Context` across boundaries and external calls.
- Use small, focused interfaces defined in consumer packages.
- Prefer early returns and avoid panics for normal control flow.

### golang-testing
- Follow RED → GREEN → REFACTOR for every behavior change.
- Use table-driven tests and subtests as default structure.
- Use `t.Helper()` and `t.Cleanup()` in reusable test helpers.
- Favor deterministic tests (no sleeps, no hidden shared state).
- Use `httptest` for HTTP handlers and boundary contracts.
- Track coverage with `go test -cover`/`-coverprofile` and close gaps.

### issue-creation
- Use issue templates only; blank issues are not valid workflow.
- New issues start with `status:needs-review`; PRs require `status:approved`.
- Route questions to Discussions, not Issues.
- Ensure required template fields are complete and reproducible.
- Search for duplicates before creating a new issue.
- Keep issue titles in conventional style (`fix(scope): ...`, `feat(scope): ...`).

### judgment-day
- Launch two blind judges in parallel; never sequentially.
- Resolve and inject relevant compact rules into all judge/fix prompts.
- Treat only confirmed CRITICAL/real WARNING as mandatory blockers.
- After fixes, re-judge immediately; do not conclude before verdict.
- Ask user before continuing beyond two fix iterations.
- Return explicit terminal state: APPROVED or ESCALATED.

### nothing-ui-shell
- Use strict monochrome palette; avoid loud gradients and neon accents.
- Use system font stack only; maintain restrained typography hierarchy.
- Structure UI into explicit sections/tabs/sidebar; avoid mixed infinite flow.
- Keep interaction feedback subtle (soft hover/focus, gentle motion).
- Use minimal hardware cues (fine borders, dots, light textures).
- Keep visual noise low and active navigation state clear.

### postgres-patterns
- Choose index type by query pattern (B-tree, GIN, BRIN, partial/composite).
- Prefer `timestamptz`, `numeric` for money, and practical native types.
- Use cursor pagination over OFFSET for large datasets.
- Use `ON CONFLICT` for idempotent upserts.
- Optimize RLS policies and avoid unindexed FK/query hot paths.
- Track slow queries with `pg_stat_statements` and tune iteratively.

### reviewer
- Prioritize correctness, security, reliability, and architecture boundaries first.
- Focus on high-impact findings; avoid cosmetic-only feedback.
- Tie each finding to exact file/function and user impact.
- Propose actionable fixes and required tests for each risk.
- Validate error handling, cancellation/timeouts, and dependency boundaries.
- Separate must-fix blockers from high-priority improvements and nice-to-haves.

### skill-creator
- Create skills only for repeatable patterns or complex workflows.
- Use canonical `skills/{name}/SKILL.md` structure and complete frontmatter.
- Include explicit Trigger text in description for discoverability.
- Keep guidance actionable: critical patterns, short examples, concrete commands.
- Prefer local references/assets; avoid duplicating long docs in skill body.
- Register new skills in project index/conventions after creation.

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| AGENTS.md | /home/ale/Documents/Projects/micha/AGENTS.md | Index — references files below |
| cmd/api/main.go | /home/ale/Documents/Projects/micha/backend/cmd/api/main.go | Referenced by AGENTS.md |
| internal/adapters/http/server.go | /home/ale/Documents/Projects/micha/backend/internal/adapters/http/server.go | Referenced by AGENTS.md |
| internal/adapters/http/*_handler.go | /home/ale/Documents/Projects/micha/backend/internal/adapters/http/*_handler.go | Referenced by AGENTS.md |
| internal/adapters/postgres/*_repository.go | /home/ale/Documents/Projects/micha/backend/internal/adapters/postgres/*_repository.go | Referenced by AGENTS.md |
| internal/domain/shared/errors.go | /home/ale/Documents/Projects/micha/backend/internal/domain/shared/errors.go | Referenced by AGENTS.md |
| internal/adapters/ | /home/ale/Documents/Projects/micha/backend/internal/adapters/ | Referenced by AGENTS.md |
| internal/ports/ | /home/ale/Documents/Projects/micha/backend/internal/ports/ | Referenced by AGENTS.md |
| internal/application/ | /home/ale/Documents/Projects/micha/backend/internal/application/ | Referenced by AGENTS.md |
| internal/domain/ | /home/ale/Documents/Projects/micha/backend/internal/domain/ | Referenced by AGENTS.md |
| internal/infrastructure/ | /home/ale/Documents/Projects/micha/backend/internal/infrastructure/ | Referenced by AGENTS.md |
| docs/architecture-checklist.md | /home/ale/Documents/Projects/micha/docs/architecture-checklist.md | Referenced by AGENTS.md |
| backend/migrations/ | /home/ale/Documents/Projects/micha/backend/migrations/ | Referenced by AGENTS.md |
| deploy/docker-compose.yml | /home/ale/Documents/Projects/micha/deploy/docker-compose.yml | Referenced by AGENTS.md |

Read the convention files listed above for project-specific patterns and rules. All referenced paths have been extracted — no need to read index files to discover more.
