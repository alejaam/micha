# micha — Claude Code Instructions

Go REST API for personal expense tracking. Backend: Go 1.23, PostgreSQL.
Architecture: DDD + Clean Architecture + Hexagonal.

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

## Feature order (strict)

```
domain → use case → ports → adapter → wiring in cmd/api
```

Validate with `docs/architecture-checklist.md` before finishing.

## Build & test

```bash
cd backend && go run ./cmd/api
cd backend && go test ./...
cd backend && go test -race ./...
docker compose -f deploy/docker-compose.yml up --build
```

## Current state

- PostgreSQL: `localhost:5432`, db/user/pass: `micha` / `micha` / `micha_dev_password`
- Endpoints: `/health` + full CRUD (expenses, households, members) + settlement calculation
- `expense_repository.go` is a stub; migrations live in `backend/migrations/`

## Go coding standards

- Error wrapping: `fmt.Errorf("doing X: %w", err)`
- Logging: `slog.InfoContext(ctx, "event", "key", value)` — never log PII
- SQL: parameterised queries only — no string concatenation
- HTTP: `MaxBytesReader`, `DisallowUnknownFields()`, guard clauses
- Tests: table-driven, `testify/require`, `t.Parallel()`, black-box packages
