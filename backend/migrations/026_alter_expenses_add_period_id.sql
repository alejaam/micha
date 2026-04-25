ALTER TABLE expenses ADD COLUMN period_id TEXT REFERENCES periods(id);
CREATE INDEX idx_expenses_period_id ON expenses(period_id);
