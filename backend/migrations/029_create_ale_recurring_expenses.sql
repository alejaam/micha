-- 029_create_ale_recurring_expenses.sql
-- Seeds recurring monthly expenses for user ale.jaam21@gmail.com.
--
-- These are household-level fixed expenses (is_agnostic = true).
-- No specific payer assigned — each member's contribution is calculated
-- by the household's settlement mode or split config.
--
-- Categories assigned by service type:
--   Streaming services     → streaming
--   Internet/Cable (Megacable) → streaming
--   Phone (Telcel)         → personal
--
-- Apply with: psql $DATABASE_URL -f migrations/029_create_ale_recurring_expenses.sql

DO $$
DECLARE
    v_household_id  TEXT;
    v_streaming_id  TEXT;
    v_personal_id   TEXT;
    v_today         DATE := CURRENT_DATE;
BEGIN
    -- ── Look up the household where this user is a member ──
    SELECT m.household_id INTO v_household_id
    FROM members m
    WHERE m.email = 'ale.jaam21@gmail.com'
    LIMIT 1;

    IF v_household_id IS NULL THEN
        RAISE EXCEPTION 'No household found for user ale.jaam21@gmail.com. Create the household and member first.';
    END IF;

    -- ── Look up category IDs ──
    SELECT id INTO v_streaming_id
    FROM categories
    WHERE household_id = v_household_id AND slug = 'streaming';

    SELECT id INTO v_personal_id
    FROM categories
    WHERE household_id = v_household_id AND slug = 'personal';

    IF v_streaming_id IS NULL OR v_personal_id IS NULL THEN
        RAISE EXCEPTION 'Required categories not found. Ensure default categories exist for this household.';
    END IF;

    -- ── Insert recurring expenses ──
    -- Apple One Familiar: 329.00 MXN / month
    INSERT INTO recurring_expenses (id, household_id, paid_by_member_id, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, next_generation_date, is_active, is_agnostic, created_at, updated_at)
    VALUES (gen_random_uuid()::text, v_household_id, NULL, 32900, 'Apple One Familiar', v_streaming_id, 'fixed', 'monthly', v_today, v_today, true, true, now(), now())
    ON CONFLICT DO NOTHING;

    -- YouTube / INVEX TDC: 319.00 MXN / month
    INSERT INTO recurring_expenses (id, household_id, paid_by_member_id, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, next_generation_date, is_active, is_agnostic, created_at, updated_at)
    VALUES (gen_random_uuid()::text, v_household_id, NULL, 31900, 'YouTube Premium', v_streaming_id, 'fixed', 'monthly', v_today, v_today, true, true, now(), now())
    ON CONFLICT DO NOTHING;

    -- Megacable: 749.00 MXN / month
    INSERT INTO recurring_expenses (id, household_id, paid_by_member_id, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, next_generation_date, is_active, is_agnostic, created_at, updated_at)
    VALUES (gen_random_uuid()::text, v_household_id, NULL, 74900, 'Megacable Internet + TV', v_streaming_id, 'fixed', 'monthly', v_today, v_today, true, true, now(), now())
    ON CONFLICT DO NOTHING;

    -- Vix: 124.92 MXN / month
    INSERT INTO recurring_expenses (id, household_id, paid_by_member_id, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, next_generation_date, is_active, is_agnostic, created_at, updated_at)
    VALUES (gen_random_uuid()::text, v_household_id, NULL, 12492, 'Vix Premium', v_streaming_id, 'fixed', 'monthly', v_today, v_today, true, true, now(), now())
    ON CONFLICT DO NOTHING;

    -- Meli+: 569.00 MXN / month
    INSERT INTO recurring_expenses (id, household_id, paid_by_member_id, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, next_generation_date, is_active, is_agnostic, created_at, updated_at)
    VALUES (gen_random_uuid()::text, v_household_id, NULL, 56900, 'Meli+ (Mercado Libre)', v_streaming_id, 'fixed', 'monthly', v_today, v_today, true, true, now(), now())
    ON CONFLICT DO NOTHING;

    -- Telcel: 549.00 MXN / month
    INSERT INTO recurring_expenses (id, household_id, paid_by_member_id, amount_cents, description, category_id, expense_type, recurrence_pattern, start_date, next_generation_date, is_active, is_agnostic, created_at, updated_at)
    VALUES (gen_random_uuid()::text, v_household_id, NULL, 54900, 'Telcel Plan Celular', v_personal_id, 'fixed', 'monthly', v_today, v_today, true, true, now(), now())
    ON CONFLICT DO NOTHING;

    RAISE NOTICE '✅ Recurring expenses seeded for household %', v_household_id;
END $$;
