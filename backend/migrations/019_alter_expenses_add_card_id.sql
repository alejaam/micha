-- 019_alter_expenses_add_card_id.sql
-- Add card_id foreign key to expenses table.
-- This links expenses paid with credit cards to the specific card used.

ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS card_id TEXT;

ALTER TABLE expenses
    ADD CONSTRAINT fk_expenses_card
        FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_card
    ON expenses (card_id)
    WHERE card_id IS NOT NULL AND deleted_at IS NULL;

COMMENT ON COLUMN expenses.card_id IS 'Credit card used for this expense (nullable, only for payment_method=card)';
