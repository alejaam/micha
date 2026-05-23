-- 030_create_subscription_services.sql
-- Creates the subscription_services catalog table with LATAM seed data.
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

-- Seed data: Mexico/LATAM streaming services and bundles (prices as of May 2026)
INSERT INTO subscription_services (name, slug, region, currency, standalone_price_cents, is_bundle) VALUES
    -- Netflix (individual tiers)
    ('Netflix Standard con anuncios', 'netflix-standard-anuncios', 'MX', 'MXN', 12900, false),
    ('Netflix Standard', 'netflix-standard', 'MX', 'MXN', 21900, false),
    ('Netflix Premium', 'netflix-premium', 'MX', 'MXN', 29900, false),

    -- Disney+
    ('Disney+ Standard con anuncios', 'disney-plus-anuncios', 'MX', 'MXN', 15900, false),
    ('Disney+ Premium', 'disney-plus-premium', 'MX', 'MXN', 19900, false),

    -- Max (formerly HBO Max)
    ('Max Básico con anuncios', 'max-basico-anuncios', 'MX', 'MXN', 9900, false),
    ('Max Standard', 'max-standard', 'MX', 'MXN', 14900, false),
    ('Max Premium', 'max-premium', 'MX', 'MXN', 19900, false),

    -- Apple services
    ('Apple TV+', 'apple-tv-plus', 'MX', 'MXN', 6900, false),
    ('Apple One Familiar', 'apple-one-familiar', 'MX', 'MXN', 32900, true),

    -- YouTube
    ('YouTube Premium Individual', 'youtube-premium-individual', 'MX', 'MXN', 12900, false),
    ('YouTube Premium Familiar', 'youtube-premium-familiar', 'MX', 'MXN', 31900, true),

    -- ViX
    ('ViX Premium', 'vix-premium', 'MX', 'MXN', 12400, false),

    -- Meli+ (Mercado Libre bundle)
    ('Meli+', 'meli-plus', 'MX', 'MXN', 56900, true),

    -- Spotify
    ('Spotify Individual', 'spotify-individual', 'MX', 'MXN', 12900, false),
    ('Spotify Dúo', 'spotify-duo', 'MX', 'MXN', 16900, false),
    ('Spotify Familiar', 'spotify-familiar', 'MX', 'MXN', 19900, false),

    -- Amazon
    ('Amazon Prime Video', 'amazon-prime-video', 'MX', 'MXN', 9900, false),

    -- Paramount+
    ('Paramount+', 'paramount-plus', 'MX', 'MXN', 7900, false),

    -- Megacable
    ('Megacable Internet + TV', 'megacable-internet-tv', 'MX', 'MXN', 74900, false)
ON CONFLICT DO NOTHING;
