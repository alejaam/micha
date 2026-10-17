-- Reference catalog for Mexico. Apply explicitly; schema migrations do not
-- require these rows to start the API.
DO $$
BEGIN
    IF current_setting('app.seed_scope', true) <> 'reference-mx' THEN
        RAISE EXCEPTION 'Mexico reference catalog requires SEED_SCOPE=reference-mx';
    END IF;
END $$;

INSERT INTO subscription_services (name, slug, region, currency, standalone_price_cents, is_bundle) VALUES
    ('Netflix Standard con anuncios', 'netflix-standard-anuncios', 'MX', 'MXN', 12900, false),
    ('Netflix Standard', 'netflix-standard', 'MX', 'MXN', 21900, false),
    ('Netflix Premium', 'netflix-premium', 'MX', 'MXN', 29900, false),
    ('Disney+ Standard con anuncios', 'disney-plus-anuncios', 'MX', 'MXN', 15900, false),
    ('Disney+ Premium', 'disney-plus-premium', 'MX', 'MXN', 19900, false),
    ('Max Básico con anuncios', 'max-basico-anuncios', 'MX', 'MXN', 9900, false),
    ('Max Standard', 'max-standard', 'MX', 'MXN', 14900, false),
    ('Max Premium', 'max-premium', 'MX', 'MXN', 19900, false),
    ('Apple TV+', 'apple-tv-plus', 'MX', 'MXN', 6900, false),
    ('Apple One Familiar', 'apple-one-familiar', 'MX', 'MXN', 32900, true),
    ('YouTube Premium Individual', 'youtube-premium-individual', 'MX', 'MXN', 12900, false),
    ('YouTube Premium Familiar', 'youtube-premium-familiar', 'MX', 'MXN', 31900, true),
    ('ViX Premium', 'vix-premium', 'MX', 'MXN', 12400, false),
    ('Meli+', 'meli-plus', 'MX', 'MXN', 56900, true),
    ('Spotify Individual', 'spotify-individual', 'MX', 'MXN', 12900, false),
    ('Spotify Dúo', 'spotify-duo', 'MX', 'MXN', 16900, false),
    ('Spotify Familiar', 'spotify-familiar', 'MX', 'MXN', 19900, false),
    ('Amazon Prime Video', 'amazon-prime-video', 'MX', 'MXN', 9900, false),
    ('Paramount+', 'paramount-plus', 'MX', 'MXN', 7900, false),
    ('Megacable Internet + TV', 'megacable-internet-tv', 'MX', 'MXN', 74900, false)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    region = EXCLUDED.region,
    currency = EXCLUDED.currency,
    standalone_price_cents = EXCLUDED.standalone_price_cents,
    is_bundle = EXCLUDED.is_bundle,
    updated_at = now();
