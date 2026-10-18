# Base de Datos por Ambiente (micha v1)

El arranque de la API (`backend/cmd/api`) ejecuta de forma automática las migraciones en `backend/migrations/` protegidas por un `pg_advisory_lock` y verificación de `checksum`.

Este proceso es idempotente y apto para bases de datos limpias (zero-state) en local, staging (Dokploy) y producción.

## Estructura de Migraciones

- `backend/migrations/001_initial_schema.sql`: Esquema canónico v1 (tablas, constraints, llaves foráneas e índices). Es **100% DDL estructural**. La API nunca se bloquea por datos de catálogo o fixtures.
- `backend/catalogs/reference/001_subscription_services_mx.sql`: Archivo SQL de referencia con el catálogo oficial de servicios de streaming y bundles en México (MXN) con precios actualizados.

## Arranque de la API

```bash
cd backend
DATABASE_URL=postgres://micha:micha_dev_password@localhost:5432/micha?sslmode=disable \
JWT_SECRET=dev_secret_key_at_least_32_characters_long \
ENVIRONMENT=development \
go run ./cmd/api
```

La API aplica las migraciones estructurales en el primer arranque. En arranques subsecuentes, verifica los checksums y continúa inmediatamente.

## Política de Seguridad: Sin Seeds

- **Cero seeds en código y binarios**: No existe ningún comando `seed`, fixtures ni usuarios demo en el repositorio ni en los binarios de la aplicación.
- Toda cuenta, hogar, miembro y gasto se crea exclusivamente a través del flujo normal de la aplicación (registro y onboarding vía API).
- Para cargar o actualizar el catálogo de servicios de suscripción de manera opcional y manual, se ejecuta directamente el archivo SQL mediante cualquier cliente de PostgreSQL:
  ```bash
  psql $DATABASE_URL -f backend/catalogs/reference/001_subscription_services_mx.sql
  ```

## Despliegue en Dokploy

En el servicio de Dokploy (staging / producción) configurar únicamente las variables normales de la aplicación:
- `DATABASE_URL`
- `JWT_SECRET`
- `ENVIRONMENT` (`production` o `staging`)
- `HTTP_PORT` (por defecto `8080`)
- `ALLOWED_ORIGINS`

No se requiere ningún paso manual adicional para que la API arranque y opere con normalidad.
