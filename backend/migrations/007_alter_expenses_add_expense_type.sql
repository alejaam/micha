ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS expense_type TEXT NOT NULL DEFAULT 'variable'
        CHECK (expense_type IN ('fixed', 'variable', 'msi'));
