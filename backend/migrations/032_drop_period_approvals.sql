-- Drop the period_approvals table (review/approval workflow removed).
-- The table is expected to be empty in production since no real reviews happened.
DROP TABLE IF EXISTS period_approvals;
