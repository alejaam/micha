-- 031_create_catalog_links.sql
-- Creates the many-to-many junction table between recurring expenses
-- and subscription catalog services.

CREATE TABLE IF NOT EXISTS recurring_expense_catalog_links (
    recurring_expense_id  TEXT        NOT NULL REFERENCES recurring_expenses(id) ON DELETE CASCADE,
    catalog_service_id    TEXT        NOT NULL REFERENCES subscription_services(id) ON DELETE CASCADE,
    custom_price_cents    BIGINT      CHECK (custom_price_cents IS NULL OR custom_price_cents > 0),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (recurring_expense_id, catalog_service_id)
);

CREATE INDEX IF NOT EXISTS idx_catalog_links_service
    ON recurring_expense_catalog_links (catalog_service_id);
