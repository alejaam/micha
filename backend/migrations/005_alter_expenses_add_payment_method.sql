ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS payment_method TEXT NOT NULL DEFAULT 'cash'
        CHECK (payment_method IN ('cash', 'card', 'transfer', 'voucher'));

CREATE INDEX IF NOT EXISTS idx_expenses_household_period
    ON expenses (household_id, created_at)
    WHERE deleted_at IS NULL;
