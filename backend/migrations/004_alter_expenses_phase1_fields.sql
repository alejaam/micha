ALTER TABLE expenses
    ADD COLUMN paid_by_member_id TEXT REFERENCES members(id),
    ADD COLUMN is_shared         BOOLEAN     NOT NULL DEFAULT true,
    ADD COLUMN currency          CHAR(3)     NOT NULL DEFAULT 'MXN';

CREATE INDEX idx_expenses_paid_by_member
    ON expenses (paid_by_member_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_expenses_household_shared
    ON expenses (household_id, is_shared, created_at DESC)
    WHERE deleted_at IS NULL;
