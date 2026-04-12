-- seed-flow-split.sql
-- Flujo: split proporcional entre dos miembros con sueldos distintos.
-- Alice: 50,000 MXN/mes → 62.5% del split
-- Bob:   30,000 MXN/mes → 37.5% del split
--
-- Pre-requisito: haber corrido seed.sql (o reset.sh)
-- Uso: psql $DATABASE_URL -f scripts/test-db/seed-flow-split.sql

BEGIN;

-- Actualizar sueldos para que el split sea exacto y predecible
UPDATE members SET monthly_salary_cents = 5000000 WHERE id = 'member-alice-001';
UPDATE members SET monthly_salary_cents = 3000000 WHERE id = 'member-bob-001';

-- Gastos extra para probar la pantalla de balance
-- Total shared: $5,000 MXN
-- Alice paga: $5,000 → Bob le debe $1,875 (37.5%)
INSERT INTO expenses (
  id, household_id, amount_cents, description,
  paid_by_member_id, expense_type, payment_method,
  card_id, card_name, category_id,
  created_at, updated_at
) VALUES
  ('exp-split-001', 'hh-test-001', 500000, 'Seguro del coche (split test)',
   'member-alice-001', 'shared', 'cash',
   NULL, NULL, 'cat-auto-001',
   NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),

  ('exp-split-002', 'hh-test-001', 200000, 'Internet fibra (split test)',
   'member-bob-001', 'shared', 'cash',
   NULL, NULL, 'cat-other-001',
   NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- Resumen del split esperado
SELECT
  m.name,
  m.monthly_salary_cents / 100.0 AS salary_mxn,
  ROUND(m.monthly_salary_cents * 100.0 /
    (SELECT SUM(monthly_salary_cents) FROM members WHERE household_id = 'hh-test-001' AND deleted_at IS NULL),
  2) AS split_pct
FROM members m
WHERE m.household_id = 'hh-test-001'
  AND m.deleted_at IS NULL
ORDER BY m.name;

SELECT 'seed-flow-split.sql: gastos de split insertados ✓' AS status;
