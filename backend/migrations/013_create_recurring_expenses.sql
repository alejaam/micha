-- 013_create_recurring_expenses.sql
-- Apply with: psql $DATABASE_URL -f migrations/013_create_recurring_expenses.sql

CREATE TABLE IF NOT EXISTS recurring_expenses (
    id                    TEXT        PRIMARY KEY,
    household_id          TEXT        NOT NULL,
    paid_by_member_id     TEXT        NOT NULL,
    amount_cents          BIGINT      NOT NULL CHECK (amount_cents > 0),
    description           TEXT        NOT NULL DEFAULT '',
    category_id           TEXT        NOT NULL DEFAULT '',
    expense_type          TEXT        NOT NULL DEFAULT 'fixed'
                                      CHECK (expense_type IN ('fixed', 'variable', 'msi')),
    recurrence_pattern    TEXT        NOT NULL CHECK (recurrence_pattern IN ('monthly', 'biweekly', 'weekly')),
    start_date            DATE        NOT NULL,
    end_date              DATE,
    next_generation_date  DATE        NOT NULL,
    is_active             BOOLEAN     NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at            TIMESTAMPTZ,
    
    CONSTRAINT fk_recurring_expenses_household
        FOREIGN KEY (household_id) REFERENCES households(id) ON DELETE CASCADE,
    CONSTRAINT fk_recurring_expenses_member
        FOREIGN KEY (paid_by_member_id) REFERENCES members(id) ON DELETE CASCADE,
    CONSTRAINT fk_recurring_expenses_category
        FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET DEFAULT,
    CONSTRAINT check_end_date_after_start
        CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS idx_recurring_expenses_household
    ON recurring_expenses (household_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_recurring_expenses_next_generation
    ON recurring_expenses (next_generation_date)
    WHERE deleted_at IS NULL AND is_active = true;

COMMENT ON TABLE recurring_expenses IS 'Templates for automatically generating recurring expenses (rent, subscriptions, etc.)';
COMMENT ON COLUMN recurring_expenses.recurrence_pattern IS 'Frequency: monthly, biweekly, or weekly';
COMMENT ON COLUMN recurring_expenses.next_generation_date IS 'Next date when an expense should be generated (idempotency key)';
COMMENT ON COLUMN recurring_expenses.is_active IS 'Whether this recurring expense is actively generating expenses';
