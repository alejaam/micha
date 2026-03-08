CREATE TABLE households (
    id              TEXT        PRIMARY KEY,
    name            TEXT        NOT NULL,
    settlement_mode TEXT        NOT NULL CHECK (settlement_mode IN ('equal', 'proportional')),
    currency        CHAR(3)     NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_households_created_at
    ON households (created_at DESC);
