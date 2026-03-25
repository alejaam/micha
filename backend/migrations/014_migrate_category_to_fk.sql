-- 014_migrate_category_to_fk.sql
-- Migrate expenses.category (TEXT) to expenses.category_id (FK to categories table)
-- This migration is idempotent and can be safely re-run.

-- Step 1: Add category_id column (nullable initially to allow data migration)
ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS category_id TEXT;

-- Step 2: Ensure all households have the default categories seeded
-- This handles any households created after migration 010 or before it ran
INSERT INTO categories (id, household_id, name, slug, is_default, created_at)
SELECT
    gen_random_uuid()::text,
    h.id,
    d.name,
    d.slug,
    true,
    now()
FROM households h
CROSS JOIN (
    VALUES
        ('Rent',      'rent'),
        ('Auto',      'auto'),
        ('Streaming', 'streaming'),
        ('Food',      'food'),
        ('Personal',  'personal'),
        ('Savings',   'savings'),
        ('Other',     'other')
) AS d(name, slug)
ON CONFLICT (household_id, slug) DO NOTHING;

-- Step 3: Migrate existing data ONLY if category column still exists
-- We use a DO block to conditionally run the UPDATE based on column existence
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'expenses'
          AND column_name = 'category'
    ) THEN
        -- Migrate data by matching expenses.category to categories.slug
        UPDATE expenses e
        SET category_id = c.id
        FROM categories c
        WHERE c.household_id = e.household_id
          AND c.slug = e.category
          AND e.category_id IS NULL;
    END IF;
END $$;

-- Step 4: Handle any expenses that couldn't be matched or have NULL category_id
-- Point them to the 'other' category in their household
UPDATE expenses e
SET category_id = (
    SELECT c.id
    FROM categories c
    WHERE c.household_id = e.household_id
      AND c.slug = 'other'
    LIMIT 1
)
WHERE e.category_id IS NULL;

-- Step 5: Drop the old category column and its constraint (if they still exist)
ALTER TABLE expenses
    DROP CONSTRAINT IF EXISTS expenses_category_check,
    DROP COLUMN IF EXISTS category;

-- Step 6: Make category_id NOT NULL and add FK constraint (only if not already done)
DO $$
BEGIN
    -- Set NOT NULL if column is nullable
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'expenses'
          AND column_name = 'category_id'
          AND is_nullable = 'YES'
    ) THEN
        ALTER TABLE expenses ALTER COLUMN category_id SET NOT NULL;
    END IF;

    -- Add FK constraint if it doesn't exist
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE constraint_name = 'expenses_category_id_fkey'
          AND table_name = 'expenses'
    ) THEN
        ALTER TABLE expenses
            ADD CONSTRAINT expenses_category_id_fkey
                FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Step 7: Drop old index on category and create new indexes on category_id
DROP INDEX IF EXISTS idx_expenses_household_category;
CREATE INDEX IF NOT EXISTS idx_expenses_household_category_id
    ON expenses (household_id, category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_expenses_category_id
    ON expenses (category_id) WHERE deleted_at IS NULL;
