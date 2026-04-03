-- 016_alter_expenses_add_total_installments.sql
-- Apply with: psql $DATABASE_URL -f migrations/016_alter_expenses_add_total_installments.sql

ALTER TABLE expenses
ADD COLUMN IF NOT EXISTS total_installments INT NOT NULL DEFAULT 0
CHECK (total_installments >= 0);

COMMENT ON COLUMN expenses.total_installments IS 'Number of installments for MSI (meses sin intereses) expenses. 0 for non-MSI expenses.';
