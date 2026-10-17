-- 030_create_subscription_services.sql
-- Creates the subscription_services catalog table. Reference rows are loaded
-- separately through backend/seeds/reference.
-- This table is a read-only catalog of available streaming/bundle services.
-- Prices are in MXN and may become stale; users can override per expense.

CREATE TABLE IF NOT EXISTS subscription_services (
    id                     TEXT        PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name                   TEXT        NOT NULL,
    slug                   TEXT        NOT NULL UNIQUE,
    region                 TEXT        NOT NULL DEFAULT 'MX',
    currency               TEXT        NOT NULL DEFAULT 'MXN',
    standalone_price_cents BIGINT     NOT NULL CHECK (standalone_price_cents > 0),
    icon_url               TEXT,
    is_bundle              BOOLEAN     NOT NULL DEFAULT false,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscription_services_slug ON subscription_services (slug);
