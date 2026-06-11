# AGENTS.md — micha

Go REST API for personal expense tracking. Backend: Go 1.23, PostgreSQL.
Architecture: DDD + Clean Architecture + Hexagonal. Claude Code / multi-agent compatible.

## Layer map (dependency direction: inward only)

```
cmd/api/main.go          ← composition root / wiring
internal/adapters/       ← HTTP + Postgres (implements ports)
internal/ports/          ← inbound (use-case contracts) + outbound (repo contracts)
internal/application/    ← use cases (depends on domain + ports only)
internal/domain/         ← entities, value objects, domain errors (no internal deps)
internal/infrastructure/ ← config, migrations
```

## Request flow — template for new endpoints

1. Route registered in `internal/adapters/http/server.go`
2. Handler (`internal/adapters/http/*_handler.go`) decodes request → calls inbound port
3. Use case (`internal/application/*/`) creates domain entity via constructor → calls outbound port
4. Domain validates invariants (`internal/domain/*/`)
5. Postgres adapter (`internal/adapters/postgres/*_repository.go`) fulfills outbound port

## Domain conventions

- Rich constructors: `expense.New(...)` / `expense.NewFromAttributes(...)` — never ad-hoc struct literals
- Entity fields unexported; behaviour exposed via method receivers
- Rehydration from DB uses `*Attributes` structs
- Shared domain errors: `internal/domain/shared/errors.go`
- Packages short/lowercase: `httpadapter`, `expenseapp`, `config`

## Feature order (strict — do not skip layers)

```
domain → use case → ports → adapter → wiring in cmd/api
```

Validate with `docs/architecture-checklist.md` before finishing any feature.

## Agent Routing (OpenCode)

| Task | Agent / Command |
|------|-----------------|
| Implement a feature | `build` (default) |
| Plan a feature across layers | `/plan-feature` → `plan` agent |
| Audit architecture | `/check-arch` → `architect` subagent |
| Review code quality | `/review` → `reviewer` subagent |
| Run & fix tests | `/test` → `build` agent |

## Build & test

```bash
cd backend && go run ./cmd/api              # start API (port 8080)
cd backend && go test ./...                 # all tests
cd backend && go test -race ./...           # with race detector
docker compose -f deploy/docker-compose.yml up --build  # full stack (Go + Postgres)
```

## Current state

- `expense_repository.go` stub (`Save` returns `nil`); real SQL in `backend/migrations/`
- Endpoints: `/health` + full CRUD for expenses, households, members + settlement calculation
- PostgreSQL: `localhost:5432`, db/user/pass: `micha` / `micha` / `micha_dev_password`

## Agent skills

### Issue tracker

GitHub Issues for `alejaam/micha`. See `docs/agents/issue-tracker.md`.

### Triage labels

Default canonical labels (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### CodeGraph

`.codegraph/` exists in this project — a pre-indexed semantic knowledge graph of the codebase.

**Use CodeGraph tools instead of grep/read loops.** Answer directly from the graph in 2-3 calls. The returned source is authoritative — don't re-read those files with Read/Grep.

- `codegraph_context` — Map a task/feature/area first (combines search + node + callers + callees)
- `codegraph_trace` — Trace call path between symbols ("how does X reach Y")
- `codegraph_explore` — Survey several related symbols' source in one call
- `codegraph_search` — Find a symbol by name
- `codegraph_callers` / `codegraph_callees` — Walk call flow one hop at a time
- `codegraph_impact` — Check what's affected before editing
- `codegraph_node` — Get a single symbol's source/signature
- `codegraph_files` — File structure navigation
- `codegraph_status` — Index health

### Domain docs

Multi-context repo (`backend/`, `frontend/`). See `docs/agents/domain.md`.

