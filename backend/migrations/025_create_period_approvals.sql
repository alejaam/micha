CREATE TABLE period_approvals (
    id TEXT PRIMARY KEY,
    member_id TEXT NOT NULL REFERENCES members(id),
    period_id TEXT NOT NULL REFERENCES periods(id),
    status TEXT NOT NULL CHECK (status IN ('approved', 'objected')),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(member_id, period_id)
);

CREATE INDEX idx_period_approvals_period_id ON period_approvals(period_id);
