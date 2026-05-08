ALTER TABLE households
ADD COLUMN closing_day INT NOT NULL DEFAULT 15 CHECK (closing_day >= 1 AND closing_day <= 31),
ADD COLUMN period_frequency TEXT NOT NULL DEFAULT 'monthly' CHECK (period_frequency IN ('monthly', 'biweekly'));
