#!/usr/bin/env bash
# reset.sh — Reset completo de la DB de test para micha
# Uso: bash scripts/test-db/reset.sh [--skip-seed]
#
# Pasos:
#   1. Verifica que NO estás apuntando a staging
#   2. Termina conexiones activas
#   3. Drop + Create de la base de datos
#   4. Aplica todas las migraciones en orden
#   5. Siembra datos dummy base

set -euo pipefail

# ──────────────────────────────────────────────
# Config
# ──────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"\ && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/../.."\ && pwd)"
MIGRATIONS_DIR="$BACKEND_DIR/migrations"

# Cargar .env si existe
if [ -f "$BACKEND_DIR/.env" ]; then
  export $(grep -v '^#' "$BACKEND_DIR/.env" | xargs)
fi

if [ -z "${DATABASE_URL:-}" ]; then
  echo "❌  DATABASE_URL no está definida. Exporta la variable o crea backend/.env"
  exit 1
fi

# ──────────────────────────────────────────────
# Guardia anti-staging
# ──────────────────────────────────────────────
if echo "$DATABASE_URL" | grep -qvE 'localhost|127\.0\.0\.1|0\.0\.0\.0'; then
  echo "⚠️   La DATABASE_URL no parece apuntar a localhost:"
  echo "     $DATABASE_URL"
  read -r -p "   ¿Seguro que quieres continuar? Escribe YES para confirmar: " confirm
  if [ "$confirm" != "YES" ]; then
    echo "Abortado."
    exit 1
  fi
fi

# Extraer nombre de la DB del URL
DB_NAME=$(echo "$DATABASE_URL" | sed -n 's|.*/\([^?]*\).*|\1|p')
DB_HOST_PORT=$(echo "$DATABASE_URL" | sed -n 's|postgres://[^@]*@\([^/]*\)/.*|\1|p')
DB_USER=$(echo "$DATABASE_URL" | sed -n 's|postgres://\([^:]*\):.*|\1|p')

echo ""
echo "🔄  Reseteando DB: $DB_NAME @ $DB_HOST_PORT"
echo "────────────────────────────────────────"

# ──────────────────────────────────────────────
# 1. Terminar conexiones activas
# ──────────────────────────────────────────────
echo "[1/4] Terminando conexiones activas..."
psql "$DATABASE_URL" -c \
  "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();" \
  > /dev/null 2>&1 || true

# ──────────────────────────────────────────────
# 2. Drop y re-crear la DB
# ──────────────────────────────────────────────
echo "[2/4] Dropeando y re-creando la DB..."
POSTGRES_URL=$(echo "$DATABASE_URL" | sed "s|/$DB_NAME|/postgres|g")
psql "$POSTGRES_URL" -c "DROP DATABASE IF EXISTS \"$DB_NAME\";"
psql "$POSTGRES_URL" -c "CREATE DATABASE \"$DB_NAME\" OWNER $DB_USER;"

# ──────────────────────────────────────────────
# 3. Aplicar migraciones
# ──────────────────────────────────────────────
echo "[3/4] Aplicando migraciones..."
for migration in "$MIGRATIONS_DIR"/*.sql; do
  echo "   → $(basename "$migration")"
  psql "$DATABASE_URL" -f "$migration"
done

# ──────────────────────────────────────────────
# 4. Seed
# ──────────────────────────────────────────────
SKIP_SEED=${1:-""}
if [ "$SKIP_SEED" != "--skip-seed" ]; then
  echo "[4/4] Sembrando datos dummy base..."
  psql "$DATABASE_URL" -f "$SCRIPT_DIR/seed.sql"
else
  echo "[4/4] Seed omitido (--skip-seed)"
fi

echo ""
echo "✅  Reset completo. DB lista para testing."
echo "   DB:      $DB_NAME"
echo "   Host:    $DB_HOST_PORT"
echo ""
echo "Usuarios de test:"
echo "   alice@test.micha / test1234"
echo "   bob@test.micha   / test1234"
echo "   carol@test.micha / test1234  (sin household, para flujo invitaciones)"
