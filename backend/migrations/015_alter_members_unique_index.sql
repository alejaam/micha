-- Fix for duplicate member constraint: only apply UNIQUE to non-deleted members
ALTER TABLE members DROP CONSTRAINT IF EXISTS members_household_id_email_key;
CREATE UNIQUE INDEX idx_members_household_email_active 
    ON members (household_id, email) WHERE deleted_at IS NULL;
