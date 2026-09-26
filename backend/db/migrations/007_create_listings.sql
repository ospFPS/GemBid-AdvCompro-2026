CREATE TABLE public.listings (
    listing_id BIGSERIAL PRIMARY KEY,
    gem_id BIGINT NOT NULL
        REFERENCES public.gems(gem_id)
        ON DELETE RESTRICT,
    seller_id BIGINT NOT NULL
        REFERENCES public.seller_profiles(seller_id)
        ON DELETE RESTRICT,
    sale_type VARCHAR(20)
        CHECK (sale_type IS NULL OR sale_type IN ('FIXED_PRICE','AUCTION')),
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT'
        CHECK (status IN ('DRAFT','PUBLISHED','ACTIVE','SOLD','UNSOLD','EXPIRED','CANCELLED','SUSPENDED')),
    fixed_price NUMERIC(12,2)
        CHECK (fixed_price IS NULL OR fixed_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    relisted_from_listing_id BIGINT
        REFERENCES public.listings(listing_id)
        ON DELETE SET NULL,
    CHECK (
        (sale_type IS NULL AND status = 'DRAFT' AND fixed_price IS NULL)
        OR (sale_type = 'FIXED_PRICE' AND fixed_price IS NOT NULL)
        OR (sale_type = 'AUCTION' AND fixed_price IS NULL)
    )
);
