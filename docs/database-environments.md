# Base de Datos por Ambiente (micha v1)

El arranque de la API (`backend/cmd/api`) ejecuta de forma automática las migraciones en `backend/migrations` protegidas por un `pg_advisory_lock` y verificación de `checksum`.

Este proceso es idempotente y apto para bases de datos limpias (zero-state) en local, staging (Dokploy) y producción.

## Estructura de Migraciones

- `001_initial_schema.sql`: Esquema canónico v1 (tablas, constraints, llaves foráneas e índices).
- `002_subscription_services.sql`: Catálogo de referencia de servicios de streaming y bundles en México (MXN) con precios actualizados.

## Arranque Normal

```bash
cd backend
DATABASE_URL=postgres://micha:micha_dev_password@localhost:5432/micha?sslmode=disable \
JWT_SECRET=dev_secret_key \
ENVIRONMENT=development \
go run ./cmd/api
```

La API aplica las migraciones estructurales y el catálogo base en el primer arranque. En arranques subsecuentes, verifica los checksums y continúa inmediatamente.

## Datos de Prueba en Desarrollo (Opcional)

Para poblar rápidamente un entorno local de desarrollo con datos de prueba realistas (un usuario demo, un hogar demo con miembros, salarios, gastos y servicios recurrentes):

```bash
cd backend
DATABASE_URL=postgres://micha:micha_dev_password@localhost:5432/micha?sslmode=disable \
ENVIRONMENT=development \
go run ./cmd/seed
```

Credenciales de prueba generadas:
- **Usuario**: `demo@micha.local`
- **Contraseña**: `password123`

> [!NOTE]
> `cmd/seed` es una herramienta exclusiva de desarrollo. No se registra en `schema_migrations` y está terminantemente bloqueada en entornos `production` y `prod`.

## Despliegue en Dokploy

En el servicio de Dokploy (staging / producción) configurar únicamente las variables normales de la aplicación:
- `DATABASE_URL`
- `JWT_SECRET`
- `ENVIRONMENT` (`production` o `staging`)
- `HTTP_PORT` (por defecto `8080`)
- `ALLOWED_ORIGINS`

No se requiere ningún job manual adicional para disponer del catálogo ni de la estructura completa de base de datos.
