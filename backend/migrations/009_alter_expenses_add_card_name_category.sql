-- 009: add card_name and category to expenses
-- card_name: free-text label for the card used (e.g. "BANAMEX", "HSBC AIR", "BBVA")
-- category: semantic grouping matching the Excel panels
--   rent | auto | streaming | food | personal | savings | other

ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS card_name TEXT    NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS category  TEXT    NOT NULL DEFAULT 'other'
        CHECK (category IN ('rent', 'auto', 'streaming', 'food', 'personal', 'savings', 'other'));

CREATE INDEX IF NOT EXISTS idx_expenses_household_category
    ON expenses (household_id, category) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_expenses_household_card
    ON expenses (household_id, card_name) WHERE deleted_at IS NULL AND card_name <> '';
