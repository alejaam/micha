-- drop-recreate.sql
-- Borra y re-crea la DB desde psql conectado a postgres (no a micha).
-- Uso (conectado a DB 'postgres'):
--   psql postgres://micha:micha_dev_password@localhost:5432/postgres -f scripts/test-db/drop-recreate.sql
--
-- ⚠️  NUNCA ejecutar contra staging.

\echo '>>> Terminando conexiones activas a micha_test...'
SELECT pg_terminate_backend(pid)
FROM   pg_stat_activity
WHERE  datname = 'micha_test'
  AND  pid <> pg_backend_pid();

\echo '>>> Dropeando DB...'
DROP DATABASE IF EXISTS micha_test;

\echo '>>> Creando DB...'
CREATE DATABASE micha_test OWNER micha;

\echo '>>> Listo. Ahora corre las migraciones con reset.sh o manualmente.'
