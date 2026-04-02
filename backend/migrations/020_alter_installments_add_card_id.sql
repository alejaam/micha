-- 020_alter_installments_add_card_id.sql
-- Add card_id foreign key to installments table.
-- Installments inherit the card from their parent MSI expense.

ALTER TABLE installments
    ADD COLUMN IF NOT EXISTS card_id TEXT;

ALTER TABLE installments
    ADD CONSTRAINT fk_installments_card
        FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_installments_card
    ON installments (card_id)
    WHERE card_id IS NOT NULL;

COMMENT ON COLUMN installments.card_id IS 'Credit card associated with the parent MSI expense (inherited from expenses.card_id)';
