-- 002: Optional "a cargo de" member on expenses.
-- An expense with charged_to_member_id is owed in full by that member and is
-- settled 1:1 with the payer instead of being split by the household mode.
-- NULL keeps the default shared behaviour.

ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS charged_to_member_id TEXT;

ALTER TABLE expenses
    ADD CONSTRAINT fk_expenses_charged_to_member_same_household
        FOREIGN KEY (household_id, charged_to_member_id)
        REFERENCES members (household_id, id)
        ON DELETE SET NULL (charged_to_member_id);

CREATE INDEX IF NOT EXISTS idx_expenses_charged_to_member
    ON expenses(charged_to_member_id)
    WHERE charged_to_member_id IS NOT NULL AND deleted_at IS NULL;
