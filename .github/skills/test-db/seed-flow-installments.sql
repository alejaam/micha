-- seed-flow-installments.sql
-- Flujo: gasto en MSI (meses sin intereses) con installments.
-- Caso: Laptop $24,000 MXN a 12 MSI en BBVA.
-- Genera la expense padre + 12 installments mensuales.
--
-- Pre-requisito: seed.sql (o reset.sh)
-- Uso: psql $DATABASE_URL -f scripts/test-db/seed-flow-installments.sql

BEGIN;

-- Gasto padre (compra en MSI)
INSERT INTO expenses (
  id, household_id, amount_cents, description,
  paid_by_member_id, expense_type, payment_method,
  card_id, card_name, category_id,
  total_installments,
  created_at, updated_at
) VALUES
  ('exp-msi-001', 'hh-test-001', 2400000, 'Laptop Dell XPS 12 MSI',
   'member-alice-001', 'shared', 'credit_card',
   'card-bbva-001', 'BBVA Azul', 'cat-personal-001',
   12,
   NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days')
ON CONFLICT (id) DO NOTHING;

-- 12 installments (una por mes, $2,000 MXN c/u)
INSERT INTO installments (
  id, expense_id, household_id, card_id,
  amount_cents, installment_number, total_installments,
  due_date, paid_at,
  created_at, updated_at
)
SELECT
  'inst-msi-' || LPAD(n::text, 3, '0'),
  'exp-msi-001',
  'hh-test-001',
  'card-bbva-001',
  200000,                              -- $2,000 MXN por cuota
  n,
  12,
  (NOW() - INTERVAL '60 days' + (n - 1) * INTERVAL '30 days')::date,
  CASE WHEN n <= 2 THEN NOW() - INTERVAL '30 days' ELSE NULL END,  -- las 2 primeras ya pagadas
  NOW() - INTERVAL '60 days',
  NOW() - INTERVAL '60 days'
FROM generate_series(1, 12) AS n
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- Resumen
SELECT
  i.installment_number,
  i.amount_cents / 100.0 AS amount_mxn,
  i.due_date,
  CASE WHEN i.paid_at IS NOT NULL THEN '✓ Pagada' ELSE '⏳ Pendiente' END AS estado
FROM installments i
WHERE i.expense_id = 'exp-msi-001'
ORDER BY i.installment_number;

SELECT 'seed-flow-installments.sql: MSI 12 meses insertados ✓' AS status;
