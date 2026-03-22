CREATE TABLE members (
    id                   TEXT        PRIMARY KEY,
    household_id         TEXT        NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    name                 TEXT        NOT NULL,
    email                TEXT        NOT NULL,
    monthly_salary_cents BIGINT      NOT NULL CHECK (monthly_salary_cents >= 0),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (household_id, email)
);

CREATE INDEX idx_members_household_created
    ON members (household_id, created_at DESC);
