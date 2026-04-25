-- 023_alter_cards_add_owner_member_id.sql
-- Add optional owner_member_id to cards so cards can be member-scoped.
-- Existing rows remain NULL to preserve legacy shared cards.

ALTER TABLE cards
    ADD COLUMN IF NOT EXISTS owner_member_id TEXT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_cards_owner_member'
    ) THEN
        ALTER TABLE cards
            ADD CONSTRAINT fk_cards_owner_member
            FOREIGN KEY (owner_member_id) REFERENCES members(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_cards_owner_member
    ON cards (owner_member_id)
    WHERE owner_member_id IS NOT NULL;

COMMENT ON COLUMN cards.owner_member_id IS 'Optional owner member for personal cards; NULL means legacy shared card';
