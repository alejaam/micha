-- 018_create_cards.sql
-- Create the cards table to track credit cards used for shared expenses.
-- Each card belongs to a household and has a monthly cutoff day.

CREATE TABLE IF NOT EXISTS cards (
    id           TEXT        PRIMARY KEY,
    household_id TEXT        NOT NULL,
    bank_name    TEXT        NOT NULL,
    card_name    TEXT        NOT NULL,
    cutoff_day   INTEGER     NOT NULL CHECK (cutoff_day >= 1 AND cutoff_day <= 31),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,

    CONSTRAINT fk_cards_household
        FOREIGN KEY (household_id) REFERENCES households(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cards_household
    ON cards (household_id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE cards IS 'Credit cards used for shared expenses within a household';
COMMENT ON COLUMN cards.bank_name IS 'Issuing bank or fintech (e.g., BANAMEX, BBVA, Nu, Rappi)';
COMMENT ON COLUMN cards.card_name IS 'User-friendly label for the card (e.g., "BBVA Azul", "Nu Mexico")';
COMMENT ON COLUMN cards.cutoff_day IS 'Day of the month (1-31) when the billing cycle closes';
