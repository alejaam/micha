-- clean.sql
-- Limpia SOLO los datos, respetando la estructura y las migraciones aplicadas.
-- Más rápido que un reset completo cuando solo quieres datos frescos.
-- Uso: psql $DATABASE_URL -f scripts/test-db/clean.sql

BEGIN;

-- Desactivar FK checks temporalmente para limpiar en cualquier orden
SET session_replication_role = replica;

TRUNCATE TABLE
    member_invitations,
    installments,
    recurring_expenses,
    expenses,
    cards,
    categories,
    household_split_config,
    members,
    households,
    users
RESTART IDENTITY CASCADE;

SET session_replication_role = DEFAULT;

COMMIT;

SELECT 'clean.sql: todas las tablas vaciadas ✓' AS status;
