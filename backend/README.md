# Micha Backend - Expense Tracking API

REST API backend for personal expense tracking and household settlement calculation.

## Tech Stack

- **Go 1.24.x**
- **PostgreSQL 15+**
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
- PostgreSQL 15 or higher
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
APP_ENV=development
```

### 2. Database Setup

Run migrations using your preferred tool. Example with `golang-migrate`:

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

Or use the provided Docker Compose setup (includes Postgres):

```bash
cd ../deploy
docker-compose up -d postgres
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
| GET | `/v1/households/{id}/settlement?month=YYYY-MM` | Calculate monthly settlement | JWT + Member |

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

1. `001_create_expenses.sql` - Initial expense table
2. `002_create_households.sql` - Household table
3. `003_create_members.sql` - Member table
4. `004_alter_expenses_phase1_fields.sql` - Add expense fields
5. `005_alter_expenses_add_payment_method.sql` - Payment methods
6. `006_create_users.sql` - User authentication
7. `007_alter_expenses_add_expense_type.sql` - Expense types
8. `008_alter_members_add_user_id.sql` - Link members to users
9. `009_alter_expenses_add_card_name_category.sql` - Card name + category
10. `010_categories.sql` - Custom categories table
11. `011_household_split_config.sql` - Split configuration
12. `012_add_deleted_at_members.sql` - Soft delete for members

## Docker Deployment

See `/deploy/docker-compose.yml` for full stack deployment:

```bash
cd ../deploy
docker-compose up --build
```

This starts:
- PostgreSQL database
- Backend API on port 8080

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HTTP_PORT` | No | `8080` | HTTP server port |
| `DATABASE_URL` | **Yes** | - | PostgreSQL connection string |
| `JWT_SECRET` | **Yes** | - | Secret key for JWT signing |
| `ALLOWED_ORIGINS` | Condicional | `*` fuera de producción | CORS allowed origins (comma-separated). Required in production and wildcard is rejected. |
| `APP_ENV` / `ENV` | No | - | Runtime environment. Use `production` or `prod` to enforce production config rules. |

### Production Safety Rules

- `JWT_SECRET` must be at least 32 characters.
- When `APP_ENV` or `ENV` is `production`/`prod`:
  - `ALLOWED_ORIGINS` is required.
  - `ALLOWED_ORIGINS` cannot contain `*`.

## Project Status

**Version**: 1.0.0-rc  
**Status**: Release Candidate (backend)

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
