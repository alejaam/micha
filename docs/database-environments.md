# Base de Datos por Ambiente (micha v1)

El arranque de la API (`backend/cmd/api`) ejecuta de forma automática las migraciones en `backend/migrations/` protegidas por un `pg_advisory_lock` y verificación de `checksum`.

Este proceso es idempotente y apto para bases de datos limpias (zero-state) en local, staging (Dokploy) y producción.

## Estructura de Migraciones y Catálogos

- `backend/migrations/001_initial_schema.sql`: Esquema canónico v1 (tablas, constraints, llaves foráneas e índices). Es **100% DDL estructural**. La API nunca se bloquea por datos de catálogo o fixtures.
- `backend/catalogs/reference/001_subscription_services_mx.sql`: Catálogo oficial de referencia de servicios de streaming y bundles en México (MXN) con precios actualizados.

## Arranque Normal de la API

```bash
cd backend
DATABASE_URL=postgres://micha:micha_dev_password@localhost:5432/micha?sslmode=disable \
JWT_SECRET=dev_secret_key_at_least_32_characters_long \
ENVIRONMENT=development \
go run ./cmd/api
```

La API aplica las migraciones estructurales en el primer arranque. En arranques subsecuentes, verifica los checksums y continúa inmediatamente.

## Carga del Catálogo de Referencia (Explícito)

Para poblar o actualizar el catálogo de servicios de suscripción de forma explícita e idempotente:

```bash
cd backend
SEED_CONFIRM=true \
SEED_KIND=reference \
DATABASE_URL=postgres://micha:micha_dev_password@localhost:5432/micha?sslmode=disable \
go run ./cmd/seed
```

Este comando:
- Requiere `SEED_CONFIRM=true` y `SEED_KIND=reference`.
- No toca ni contamina `schema_migrations`.
- Ejecuta `ON CONFLICT (slug) DO UPDATE`.

## Política de Fixtures y Seguridad

- **Cero fixtures de prueba en el repositorio**: No existen credenciales, usuarios ficticios ni datos de prueba mockeados en el código ni en los binarios de la aplicación.
- Toda cuenta, hogar y miembro se crea a través del flujo normal de la aplicación y sus casos de uso.
- `cmd/seed` rechaza cualquier tipo distinto a `reference`.

## Despliegue en Dokploy

En el servicio de Dokploy (staging / producción) configurar únicamente las variables normales de la aplicación:
- `DATABASE_URL`
- `JWT_SECRET`
- `ENVIRONMENT` (`production` o `staging`)
- `HTTP_PORT` (por defecto `8080`)
- `ALLOWED_ORIGINS`

Si un entorno de staging o producción requiere el catálogo de referencia, ejecutar un job separado con `SEED_CONFIRM=true SEED_KIND=reference`.
