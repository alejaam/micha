-- 001_create_expenses.sql
-- Apply with: psql $DATABASE_URL -f migrations/001_create_expenses.sql

CREATE TABLE IF NOT EXISTS expenses (
    id            TEXT        PRIMARY KEY,
    household_id  TEXT        NOT NULL,
    amount_cents  BIGINT      NOT NULL CHECK (amount_cents > 0),
    description   TEXT        NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_expenses_household_created
    ON expenses (household_id, created_at DESC)
    WHERE deleted_at IS NULL;
