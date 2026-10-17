# Database por ambiente

El arranque normal de `backend/cmd/api` ejecuta únicamente `backend/migrations`.
Ese directorio contiene cambios estructurales y el proceso debe funcionar contra
una base vacía: no crea usuarios, hogares, fixtures ni datos personales.

## Arranque normal

```bash
cd backend
DATABASE_URL=... JWT_SECRET=... ENVIRONMENT=development go run ./cmd/api
```

El runner usa un lock advisory de PostgreSQL y guarda checksum de cada migration.
No se deben editar archivos ya aplicados; los cambios posteriores se agregan como
nuevas migrations.

## Seeds explícitos

Los seeds requieren confirmación y scope explícitos. Nunca se ejecutan desde el
arranque de la API.

Catálogo de referencia:

```bash
cd backend
SEED_CONFIRM=true \
SEED_ENVIRONMENT=staging \
SEED_SCOPE=reference-mx \
SEED_KIND=reference \
DATABASE_URL=... \
go run ./cmd/seed
```

Fixture de desarrollo:

```bash
cd backend
SEED_CONFIRM=true \
SEED_ENVIRONMENT=local \
SEED_SCOPE=household-demo \
SEED_KIND=development \
SEED_EMAIL=usuario@example.com \
DATABASE_URL=... \
go run ./cmd/seed
```

Los fixtures están bloqueados para `prod` y `production`. El scope queda
registrado en la operación y debe ser elegido por el operador; el seed no debe
usarse como mecanismo para crear la primera cuenta o el primer hogar.

## Dokploy

Configurar en el servicio de la API solo `DATABASE_URL`, `JWT_SECRET`,
`ENVIRONMENT` y las variables normales de aplicación. No configurar
`SEED_CONFIRM` ni `SEED_KIND` en el proceso web.

Si staging necesita el catálogo o fixtures, ejecutar un job/comando separado con
las variables anteriores y con el mismo `DATABASE_URL` de staging. Producción
solo aplica migrations estructurales durante el despliegue; el catálogo de
referencia, si se desea, requiere un job explícito con `SEED_KIND=reference` y
`SEED_SCOPE=reference-mx`. Los fixtures de desarrollo están bloqueados en
producción.
