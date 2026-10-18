-- 002_subscription_services.sql
-- Reference catalog of subscription services and bundles for Mexico (MXN).
-- Populated with official verified pricing. Safe to run against clean or existing databases.

INSERT INTO subscription_services (name, slug, region, currency, standalone_price_cents, is_bundle) VALUES
    ('Netflix Estándar con anuncios', 'netflix-standard-anuncios', 'MX', 'MXN', 13900, false),
    ('Netflix Estándar', 'netflix-standard', 'MX', 'MXN', 26900, false),
    ('Netflix Premium', 'netflix-premium', 'MX', 'MXN', 36900, false),
    ('Disney+ Estándar', 'disney-plus-estandar', 'MX', 'MXN', 24900, false),
    ('Disney+ Premium', 'disney-plus-premium', 'MX', 'MXN', 31900, false),
    ('Max Básico con anuncios', 'max-basico-anuncios', 'MX', 'MXN', 14900, false),
    ('Max Estándar', 'max-standard', 'MX', 'MXN', 23900, false),
    ('Max Platino', 'max-platino', 'MX', 'MXN', 29900, false),
    ('Apple TV+', 'apple-tv-plus', 'MX', 'MXN', 16900, false),
    ('Apple One Individual', 'apple-one-individual', 'MX', 'MXN', 24900, true),
    ('Apple One Familiar', 'apple-one-familiar', 'MX', 'MXN', 32900, true),
    ('Apple One Premium', 'apple-one-premium', 'MX', 'MXN', 49900, true),
    ('YouTube Premium Individual', 'youtube-premium-individual', 'MX', 'MXN', 15900, false),
    ('YouTube Premium Familiar', 'youtube-premium-familiar', 'MX', 'MXN', 31900, true),
    ('Spotify Individual', 'spotify-individual', 'MX', 'MXN', 13900, false),
    ('Spotify Dúo', 'spotify-duo', 'MX', 'MXN', 18900, false),
    ('Spotify Familiar', 'spotify-familiar', 'MX', 'MXN', 23900, false),
    ('Amazon Prime Video', 'amazon-prime-video', 'MX', 'MXN', 12900, true),
    ('ViX Premium', 'vix-premium', 'MX', 'MXN', 14900, false),
    ('Paramount+ Estándar', 'paramount-plus-estandar', 'MX', 'MXN', 14900, false),
    ('Paramount+ Premium', 'paramount-plus-premium', 'MX', 'MXN', 19900, false),
    ('Meli+ Total', 'meli-plus-total', 'MX', 'MXN', 12900, true)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    region = EXCLUDED.region,
    currency = EXCLUDED.currency,
    standalone_price_cents = EXCLUDED.standalone_price_cents,
    is_bundle = EXCLUDED.is_bundle,
    updated_at = now();
