-- seed-flow-invitations.sql
-- Flujo: carol@test.micha recibe invitación al household de Alice.
-- Token de invitación fijo para poder testearlo directamente en la app.
--
-- Pre-requisito: seed.sql (o reset.sh)
-- Uso: psql $DATABASE_URL -f scripts/test-db/seed-flow-invitations.sql

BEGIN;

-- Asegurarse que Carol existe sin household
-- (ya fue creada por seed.sql)

-- Invitación activa con token fijo
-- Token: test-invite-token-carol-001
-- Expira en 7 días desde NOW()
INSERT INTO member_invitations (
  id,
  household_id,
  invited_email,
  token,
  invited_by_member_id,
  expires_at,
  accepted_at,
  created_at
) VALUES (
  'inv-carol-001',
  'hh-test-001',
  'carol@test.micha',
  'test-invite-token-carol-001',
  'member-alice-001',
  NOW() + INTERVAL '7 days',
  NULL,                          -- NULL = aún no aceptada
  NOW() - INTERVAL '1 hour'
)
ON CONFLICT (id) DO NOTHING;

COMMIT;

SELECT
  'Invitación creada para carol@test.micha' AS info,
  token,
  expires_at,
  CASE WHEN accepted_at IS NULL THEN 'Pendiente' ELSE 'Aceptada' END AS estado
FROM member_invitations
WHERE id = 'inv-carol-001';

SELECT 'seed-flow-invitations.sql: invitación de Carol insertada ✓' AS status;
