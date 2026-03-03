# micha — Agent Instructions

Go REST API for personal expense tracking. Go 1.23, PostgreSQL 16, DDD + Clean + Hexagonal architecture.

## Dev environment

```bash
# Run API (port 8080, override with HTTP_PORT env var)
cd backend && go run ./cmd/api

# Full stack with Docker Compose (API + PostgreSQL)
docker compose -f deploy/docker-compose.yml up --build

# Run all tests
cd backend && go test ./...

# Lint (requires golangci-lint)
cd backend && golangci-lint run ./...
```

PostgreSQL: `localhost:5432` — db: `micha`, user: `micha`, password: `micha_dev_password`

## Project layout

```
backend/
  cmd/api/main.go                    composition root
  internal/domain/expense/           entity + constructors + invariants
  internal/domain/shared/errors.go   shared domain errors
  internal/application/expense/      use cases
  internal/ports/inbound/            use-case contracts
  internal/ports/outbound/           repository contracts
  internal/adapters/http/            HTTP handlers + server
  internal/adapters/postgres/        DB adapter (currently stubbed)
  internal/infrastructure/config/    env config
  migrations/                        SQL migrations (currently empty)
deploy/docker-compose.yml
docs/architecture-checklist.md
```

## Current state — important

- `internal/adapters/postgres/expense_repository.go` → `Save` returns `nil` (stub)
- `backend/migrations/` is empty — adding persistence requires a migration here
- Only `/health` endpoint exists in code

## Adding a feature

**Order**: domain → use case → ports → adapter → wiring in `cmd/api`  
**Validate**: `docs/architecture-checklist.md` before finishing

## Testing

```bash
cd backend && go test ./...                    # all tests
cd backend && go test ./internal/domain/...    # domain only
cd backend && go test -race ./...              # with race detector
```

## PR conventions

- Branch: `feature/<short-description>` or `fix/<short-description>`
- Commits: imperative mood, lowercase, no trailing period (`add expense validation`, not `Added expense validation.`)
- Run `go test ./...` and `golangci-lint run ./...` before pushing
