-- 011_household_split_config.sql
-- Stores per-member split percentages for a household.
-- An absent row means the household uses equal split (default).

CREATE TABLE IF NOT EXISTS household_split_config (
    household_id TEXT    NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    member_id    TEXT    NOT NULL REFERENCES members(id)    ON DELETE CASCADE,
    percentage   INTEGER NOT NULL CHECK (percentage > 0 AND percentage <= 100),

    PRIMARY KEY (household_id, member_id)
);

CREATE INDEX IF NOT EXISTS idx_split_config_household_id ON household_split_config (household_id);
