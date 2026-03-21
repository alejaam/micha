-- Add soft delete support for members
ALTER TABLE members ADD COLUMN deleted_at TIMESTAMPTZ;

-- Partial index for efficient queries on active members
CREATE INDEX idx_members_active ON members(household_id) WHERE deleted_at IS NULL;
