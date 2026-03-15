-- Add optional user_id link to members so the authenticated user can be
-- resolved to a member automatically when emails match.
ALTER TABLE members
    ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id) ON DELETE SET NULL;
