# Micha

Micha es una aplicación web de finanzas para gastos compartidos, pensada para **roomies, parejas y familias**. Cada grupo es un *hogar* con miembros; se registran los gastos (únicos y recurrentes), se reparten entre los miembros y se calcula la liquidación mensual de quién le debe a quién.

## Estado actual

Lo que existe hoy en el código:

- **Autenticación** con registro, login y JWT (`/v1/auth/*`).
- **Hogares y miembros**, con onboarding en el frontend (hogar, miembros, tarjetas, gastos fijos).
- **Gastos**: alta, edición, baja y listado; gastos recurrentes y generación de los del periodo.
- **Gastos «a cargo de» un miembro** como excepción en la liquidación (migración `002`).
- **Liquidación mensual** (`GET /v1/households/{id}/settlement?year=YYYY&month=M`) y configuración de reparto (`split-config`).
- **Categorías** personalizadas, **tarjetas**, **periodos** (actual, historial, simulación de cierre) y **catálogo de servicios de suscripción** con KPI.
- **Frontend** React con rutas protegidas, panel de gastos, saldos, tarjetas, gastos fijos, mensualidades y reglas.
- **Despliegue** con Docker Compose (API Go, frontend servido por Nginx y PostgreSQL).

Pendiente (ver [`docs/product-roadmap.md`](docs/product-roadmap.md)): entre otros, recuperación de contraseña, OAuth y verificación por correo. El envío de correos de invitación no está implementado en el código actual.

## Stack

| Capa | Tecnología |
|------|------------|
| Backend | Go 1.24 (`backend/go.mod`), API REST con `net/http`, arquitectura DDD + Clean + Hexagonal |
| Frontend | React 18 + Vite 5, pruebas con Vitest, lint con ESLint |
| Base de datos | PostgreSQL 16 (`docker-compose.yml`) |
| Despliegue | Docker Compose; el frontend se sirve con Nginx |

## Estructura

```
backend/     API en Go (cmd/api, internal/{domain,application,ports,adapters,infrastructure}, migrations/, catalogs/)
frontend/    SPA React + Vite (Dockerfile y nginx.conf para producción)
docs/        principios, checklist de arquitectura, roadmap y ADRs (docs/adr/)
docker-compose.yml   stack completo: backend, frontend y base de datos
Makefile     atajos de lint, pruebas y build del monorepo
```

## Requisitos

- Go 1.24 o superior
- Node.js 20 o superior (con npm)
- PostgreSQL 16, o Docker con Docker Compose

## Correr en local

### 1. Base de datos

Levanta solo PostgreSQL con Docker (usuario, contraseña y base: `micha` / `micha_dev_password` / `micha`):

```bash
docker compose up -d database
```

Las migraciones de `backend/migrations/` se aplican solas al arrancar la API; no hay que ejecutarlas a mano. Si prefieres un PostgreSQL propio, crea una base y ajusta `DATABASE_URL`.

### 2. Backend

```bash
cd backend
cp .env.example .env
```

`.env.example` ya apunta a la base del paso anterior. El backend lee variables de entorno (no carga el archivo `.env` por sí mismo), así que expórtalas antes de arrancar:

```bash
set -a; source .env; set +a
go run ./cmd/api
```

La API queda en `http://localhost:8080`; comprueba con `curl http://localhost:8080/health`.

Variables: `DATABASE_URL` y `JWT_SECRET` son obligatorias; `HTTP_PORT` (8080), `ALLOWED_ORIGINS` (`*`) y `ENVIRONMENT` (`development`) son opcionales. Detalle en [`backend/README.md`](backend/README.md).

### 3. Frontend

En otra terminal:

```bash
cd frontend
npm install
npm run dev
```

La app queda en `http://localhost:5173`; Vite redirige `/v1` y `/health` al backend en `localhost:8080`. Abre la app y crea tu cuenta desde el registro: no hay datos de ejemplo ni usuarios demo.

### Pruebas y lint

```bash
make test    # backend (go test) y frontend (vitest con cobertura)
make lint    # golangci-lint (si está instalado) y eslint
make build   # lint + test + build de ambos
```

También por separado: `cd backend && go test ./...` y `cd frontend && npm test`.

## Correr con Docker

Desde la raíz del repo, con `DATABASE_URL` y `JWT_SECRET` definidos (el compose no los trae por defecto). El host de la base dentro de la red de Compose es `database`:

```bash
export DATABASE_URL='postgres://micha:micha_dev_password@database:5432/micha?sslmode=disable'
export JWT_SECRET='cambia-esto-por-un-secreto-largo'
docker compose up --build
```

- Frontend (Nginx): `http://localhost:5173`
- API: `http://localhost:8080`
- PostgreSQL: `localhost:5432`

El frontend lee `VITE_API_URL` (por defecto `http://localhost:8080`). Para un entorno publicado, ajusta también `ALLOWED_ORIGINS` y `ENVIRONMENT`. Notas sobre bases por ambiente y Dokploy: [`docs/database-environments.md`](docs/database-environments.md).

## Documentación

- [`backend/README.md`](backend/README.md): API, endpoints, variables y migraciones.
- [`frontend/README.md`](frontend/README.md): desarrollo y build del frontend.
- [`AGENTS.md`](AGENTS.md): capas, convenciones y flujo para agregar endpoints.
- [`MICHA_VISION.md`](MICHA_VISION.md): visión del producto.
- [`docs/product-roadmap.md`](docs/product-roadmap.md): roadmap de producto.
- [`docs/adr/`](docs/adr/): decisiones de arquitectura.
- [`docs/architecture-principles.md`](docs/architecture-principles.md) y [`docs/architecture-checklist.md`](docs/architecture-checklist.md): principios y checklist.
