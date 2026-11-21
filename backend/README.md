# Micha Backend - Expense Tracking API

REST API backend for personal expense tracking and household settlement calculation.

## Tech Stack

- **Go 1.24+**
- **PostgreSQL 16**
- **Architecture**: DDD + Clean Architecture + Hexagonal (Ports & Adapters)
- **Auth**: JWT-based authentication

## Architecture

```
cmd/api/              ← composition root / wiring
internal/
  ├── adapters/       ← HTTP handlers + Postgres repositories
  ├── ports/          ← interface contracts (inbound + outbound)
  ├── application/    ← use cases (business logic)
  ├── domain/         ← entities, value objects, domain errors
  └── infrastructure/ ← config, migrations
```

**Dependency rule (strict)**: `domain → application → ports → adapters`

## Prerequisites

- Go 1.24 or higher
- PostgreSQL 16 (or use the Docker Compose service)
- Docker & Docker Compose (optional, for containerized deployment)

## Quick Start

### 1. Environment Setup

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` and set your configuration:

```bash
# Required variables
DATABASE_URL=postgres://user:password@localhost:5432/micha?sslmode=disable
JWT_SECRET=your-secure-secret-key-here

# Optional
HTTP_PORT=8080
ALLOWED_ORIGINS=http://localhost:3000
```

### 2. Database Setup

Run migrations using your preferred tool. Example with `golang-migrate`:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Or use the provided Docker Compose setup (includes Postgres):

```bash
cd ..
docker compose up -d database
```

### 3. Run the API

```bash
go run ./cmd/api
```

The API will be available at `http://localhost:8080`

Health check: `curl http://localhost:8080/health`

## API Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/auth/register` | Register new user |
| POST | `/v1/auth/login` | Login (returns JWT) |
| GET | `/v1/auth/me` | Get current user info |

### Households

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/v1/households` | Create household | JWT |
| GET | `/v1/households` | List user's households | JWT |
| GET | `/v1/households/{id}` | Get household details | JWT + Member |
| PUT | `/v1/households/{id}` | Update household | JWT + Member |

### Members

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/v1/households/{id}/members` | Add member to household | JWT + Bootstrap |
| GET | `/v1/households/{id}/members` | List household members | JWT + Member |
| PUT | `/v1/households/{id}/members/{member_id}` | Update member | JWT + Member |
| DELETE | `/v1/households/{id}/members/{member_id}` | Delete member (soft delete) | JWT + Member |

### Expenses

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/v1/expenses` | Create expense | JWT |
| GET | `/v1/expenses/{id}` | Get expense details | JWT |
| GET | `/v1/expenses?household_id={id}` | List expenses | JWT |
| PATCH | `/v1/expenses/{id}` | Update expense | JWT |
| DELETE | `/v1/expenses/{id}` | Delete expense | JWT |

### Settlement

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/v1/households/{id}/settlement?year=YYYY&month=M` | Calculate monthly settlement | JWT + Member |

### Categories

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/v1/households/{id}/categories` | Create custom category | JWT + Member |
| GET | `/v1/households/{id}/categories` | List categories | JWT + Member |
| DELETE | `/v1/households/{id}/categories/{category_id}` | Delete category | JWT + Member |

### Split Configuration

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| PUT | `/v1/households/{id}/split-config` | Update expense split config | JWT + Member |

## Development

### Run Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Specific package
go test ./internal/adapters/http/...
```

### Code Structure

- **Domain layer** (`internal/domain/*`): Pure business logic, no dependencies
  - Rich constructors: `expense.New(...)`, `expense.NewFromAttributes(...)`
  - Unexported fields, behavior via method receivers
  - Domain errors in `domain/shared/errors.go`

- **Application layer** (`internal/application/*`): Use cases
  - Depends only on domain + port interfaces
  - No HTTP/DB knowledge

- **Ports layer** (`internal/ports/*`): Interfaces
  - `inbound/`: Use case contracts (what the app exposes)
  - `outbound/`: Repository contracts (what the app needs)

- **Adapters layer** (`internal/adapters/*`): Implementation
  - `http/`: HTTP handlers, middleware, JSON serialization
  - `postgres/`: PostgreSQL repository implementations

### Migrations

Located in `/migrations`, numbered sequentially:

1. `001_initial_schema.sql` - Initial schema (households, members, expenses, users, categories, …)
2. `002_expense_charged_to_member.sql` - Optional "a cargo de" member on expenses

## Docker Deployment

See `docker-compose.yml` (repo root) for full stack deployment:

```bash
cd ..
docker compose up --build
```

This starts:
- PostgreSQL 16 database
- Backend API on port 8080
- Frontend (Nginx) on port 5173

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HTTP_PORT` | No | `8080` | HTTP server port |
| `DATABASE_URL` | **Yes** | - | PostgreSQL connection string |
| `JWT_SECRET` | **Yes** | - | Secret key for JWT signing |
| `ALLOWED_ORIGINS` | No | `*` | CORS allowed origins (comma-separated) |

## Project Status

**Version**: 1.0.0-rc  
**Status**: Release Candidate

### Completed Features

✅ User authentication (JWT)  
✅ Household management  
✅ Member management (with soft delete)  
✅ Expense CRUD operations  
✅ Monthly settlement calculation  
✅ Custom categories  
✅ Split configuration  
✅ CORS middleware  
✅ Request ID middleware  
✅ HTTP adapter tests (30 tests)  

### Architecture Compliance

All features follow strict layer dependency rules:
- Domain has zero internal dependencies
- Application depends only on domain + ports
- Adapters implement port interfaces
- Composition root wires everything in `cmd/api/main.go`

## Contributing

This is a personal project. If you find issues or have suggestions, feel free to open an issue.

## License

Private project - not licensed for public use.
