-- 022_alter_recurring_expenses_add_agnostic.sql
-- Support household-level recurring fixed expenses without a specific payer.

ALTER TABLE recurring_expenses
    ALTER COLUMN paid_by_member_id DROP NOT NULL;

ALTER TABLE recurring_expenses
    ADD COLUMN IF NOT EXISTS is_agnostic BOOLEAN NOT NULL DEFAULT false;

UPDATE recurring_expenses
SET is_agnostic = false
WHERE is_agnostic IS DISTINCT FROM false;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'check_recurring_expenses_agnostic_consistency'
    ) THEN
        ALTER TABLE recurring_expenses DROP CONSTRAINT check_recurring_expenses_agnostic_consistency;
    END IF;
END $$;

ALTER TABLE recurring_expenses
    ADD CONSTRAINT check_recurring_expenses_agnostic_consistency
    CHECK (
        (is_agnostic = true AND expense_type = 'fixed' AND paid_by_member_id IS NULL)
        OR
        (is_agnostic = false AND paid_by_member_id IS NOT NULL AND btrim(paid_by_member_id) <> '')
    );
