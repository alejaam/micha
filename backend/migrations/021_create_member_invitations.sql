CREATE TABLE IF NOT EXISTS member_invitations (
    id            TEXT        PRIMARY KEY,
    household_id  TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    member_id     TEXT        NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    email         TEXT        NOT NULL,
    invite_code   TEXT        NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    used_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_member_invitations_household_email
    ON member_invitations (household_id, email, created_at DESC);
