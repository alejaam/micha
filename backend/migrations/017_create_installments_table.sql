-- 017_create_installments_table.sql
-- Apply with: psql $DATABASE_URL -f migrations/017_create_installments_table.sql

CREATE TABLE IF NOT EXISTS installments (
    id VARCHAR(255) PRIMARY KEY,
    expense_id VARCHAR(255) NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    paid_by_member_id VARCHAR(255) NOT NULL REFERENCES members(id),
    start_date DATE NOT NULL,
    installment_amount_cents BIGINT NOT NULL CHECK (installment_amount_cents > 0),
    total_amount_cents BIGINT NOT NULL CHECK (total_amount_cents > 0),
    total_installments INT NOT NULL CHECK (total_installments > 0),
    current_installment INT NOT NULL CHECK (current_installment > 0 AND current_installment <= total_installments),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_installments_start_date ON installments(start_date);
CREATE INDEX IF NOT EXISTS idx_installments_expense_id ON installments(expense_id);
CREATE INDEX IF NOT EXISTS idx_installments_period_lookup ON installments(start_date, paid_by_member_id);

COMMENT ON TABLE installments IS 'Monthly installment payments for MSI (meses sin intereses) expenses';
COMMENT ON COLUMN installments.expense_id IS 'Foreign key to the root MSI expense';
COMMENT ON COLUMN installments.paid_by_member_id IS 'Member who paid the root expense (and is responsible for this installment)';
COMMENT ON COLUMN installments.start_date IS 'Date when this installment is due (used for period matching)';
COMMENT ON COLUMN installments.installment_amount_cents IS 'Amount for this specific installment (distributed from total)';
COMMENT ON COLUMN installments.total_amount_cents IS 'Total amount of the root MSI expense';
COMMENT ON COLUMN installments.current_installment IS '1-indexed position (e.g., 1 of 3, 2 of 3, 3 of 3)';
