-- 010_categories.sql
-- Creates the categories table and seeds default categories for all existing households.

CREATE TABLE IF NOT EXISTS categories (
    id          TEXT        NOT NULL,
    household_id TEXT       NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL,
    is_default  BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (id),
    UNIQUE (household_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_categories_household_id ON categories (household_id);

-- Seed the 7 default categories for every existing household.
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
