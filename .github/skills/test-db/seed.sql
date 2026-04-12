-- seed.sql
-- Datos dummy BASE para testing de micha.
-- Cubre: users, household, members, cards, categories (default), expenses variados.
--
-- Usuarios:
--   alice@test.micha / test1234  → miembro admin
--   bob@test.micha   / test1234  → miembro
--   carol@test.micha / test1234  → sin household (para flow de invitaciones)
--
-- IDs fijos para facilitar referencias cruzadas en pruebas.

BEGIN;

-- ────────────────────────────────────────────
-- USERS
-- password 'test1234' hasheado con bcrypt cost 10
-- ────────────────────────────────────────────
INSERT INTO users (id, email, password_hash, created_at) VALUES
  ('user-alice-001', 'alice@test.micha', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LPVaLpPRfNi', NOW() - INTERVAL '30 days'),
  ('user-bob-001',   'bob@test.micha',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LPVaLpPRfNi', NOW() - INTERVAL '30 days'),
  ('user-carol-001', 'carol@test.micha', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LPVaLpPRfNi', NOW() - INTERVAL '10 days')
ON CONFLICT (id) DO NOTHING;

-- ────────────────────────────────────────────
-- HOUSEHOLD
-- ────────────────────────────────────────────
INSERT INTO households (id, name, settlement_mode, currency, created_at, updated_at) VALUES
  ('hh-test-001', 'Casa Test', 'proportional', 'MXN', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;

-- ────────────────────────────────────────────
-- MEMBERS (Alice y Bob pertenecen al household)
-- ────────────────────────────────────────────
INSERT INTO members (id, household_id, name, email, monthly_salary_cents, user_id, created_at, updated_at) VALUES
  ('member-alice-001', 'hh-test-001', 'Alice',   'alice@test.micha', 5000000, 'user-alice-001', NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
  ('member-bob-001',   'hh-test-001', 'Bob',     'bob@test.micha',   3000000, 'user-bob-001',   NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;

-- ────────────────────────────────────────────
-- HOUSEHOLD SPLIT CONFIG (proporcional a sueldo)
-- ────────────────────────────────────────────
INSERT INTO household_split_config (household_id, updated_at)
SELECT 'hh-test-001', NOW()
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'household_split_config')
ON CONFLICT DO NOTHING;

-- ────────────────────────────────────────────
-- CATEGORIES (default para el household)
-- ────────────────────────────────────────────
INSERT INTO categories (id, household_id, name, slug, is_default, created_at) VALUES
  ('cat-rent-001',      'hh-test-001', 'Rent',      'rent',      true, NOW() - INTERVAL '30 days'),
  ('cat-auto-001',      'hh-test-001', 'Auto',      'auto',      true, NOW() - INTERVAL '30 days'),
  ('cat-streaming-001', 'hh-test-001', 'Streaming', 'streaming', true, NOW() - INTERVAL '30 days'),
  ('cat-food-001',      'hh-test-001', 'Food',      'food',      true, NOW() - INTERVAL '30 days'),
  ('cat-personal-001',  'hh-test-001', 'Personal',  'personal',  true, NOW() - INTERVAL '30 days'),
  ('cat-savings-001',   'hh-test-001', 'Savings',   'savings',   true, NOW() - INTERVAL '30 days'),
  ('cat-other-001',     'hh-test-001', 'Other',     'other',     true, NOW() - INTERVAL '30 days')
ON CONFLICT (household_id, slug) DO NOTHING;

-- ────────────────────────────────────────────
-- CARDS
-- ────────────────────────────────────────────
INSERT INTO cards (id, household_id, bank_name, card_name, cutoff_day, created_at, updated_at) VALUES
  ('card-bbva-001', 'hh-test-001', 'BBVA',       'BBVA Azul',    17, NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days'),
  ('card-nu-001',   'hh-test-001', 'Nu',          'Nu Mexico',    1,  NOW() - INTERVAL '30 days', NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;

-- ────────────────────────────────────────────
-- EXPENSES (10 gastos variados, últimos 2 meses)
-- ────────────────────────────────────────────
INSERT INTO expenses (
  id, household_id, amount_cents, description,
  paid_by_member_id, expense_type, payment_method,
  card_id, card_name, category_id,
  created_at, updated_at
) VALUES
  -- Renta (efectivo)
  ('exp-001', 'hh-test-001', 1200000, 'Renta Marzo',
   'member-alice-001', 'shared', 'cash',
   NULL, NULL, 'cat-rent-001',
   NOW() - INTERVAL '35 days', NOW() - INTERVAL '35 days'),

  -- Despensa (BBVA)
  ('exp-002', 'hh-test-001', 85000, 'Despensa Chedraui',
   'member-bob-001', 'shared', 'credit_card',
   'card-bbva-001', 'BBVA Azul', 'cat-food-001',
   NOW() - INTERVAL '28 days', NOW() - INTERVAL '28 days'),

  -- Netflix (Nu)
  ('exp-003', 'hh-test-001', 28900, 'Netflix Abril',
   'member-alice-001', 'shared', 'credit_card',
   'card-nu-001', 'Nu Mexico', 'cat-streaming-001',
   NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days'),

  -- Gas
  ('exp-004', 'hh-test-001', 45000, 'Gas Abril',
   'member-bob-001', 'shared', 'cash',
   NULL, NULL, 'cat-other-001',
   NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),

  -- Spotify (Nu)
  ('exp-005', 'hh-test-001', 9900, 'Spotify Duo',
   'member-alice-001', 'shared', 'credit_card',
   'card-nu-001', 'Nu Mexico', 'cat-streaming-001',
   NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),

  -- Uber personal Alice (no shared)
  ('exp-006', 'hh-test-001', 15000, 'Uber CDMX',
   'member-alice-001', 'personal', 'credit_card',
   'card-bbva-001', 'BBVA Azul', 'cat-personal-001',
   NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),

  -- Gasolina (shared)
  ('exp-007', 'hh-test-001', 120000, 'Gasolina',
   'member-bob-001', 'shared', 'cash',
   NULL, NULL, 'cat-auto-001',
   NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),

  -- Lunch delivery
  ('exp-008', 'hh-test-001', 32000, 'Rappi comida',
   'member-alice-001', 'shared', 'credit_card',
   'card-nu-001', 'Nu Mexico', 'cat-food-001',
   NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),

  -- Agua + luz
  ('exp-009', 'hh-test-001', 78000, 'Agua y luz',
   'member-alice-001', 'shared', 'cash',
   NULL, NULL, 'cat-other-001',
   NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),

  -- Renta Abril
  ('exp-010', 'hh-test-001', 1200000, 'Renta Abril',
   'member-alice-001', 'shared', 'cash',
   NULL, NULL, 'cat-rent-001',
   NOW() - INTERVAL '1 day',  NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

COMMIT;

SELECT 'seed.sql: datos base insertados ✓' AS status;
