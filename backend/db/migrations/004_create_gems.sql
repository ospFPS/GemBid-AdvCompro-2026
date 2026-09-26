CREATE TABLE public.gems (
    gem_id BIGSERIAL PRIMARY KEY,
    seller_id BIGINT NOT NULL
        REFERENCES public.seller_profiles(seller_id)
        ON DELETE RESTRICT,
    gem_name VARCHAR(150) NOT NULL,
    gem_type VARCHAR(100) NOT NULL,
    carat NUMERIC(10,3) CHECK (carat IS NULL OR carat > 0),
    length NUMERIC(10,2) CHECK (length IS NULL OR length > 0),
    width NUMERIC(10,2) CHECK (width IS NULL OR width > 0),
    height NUMERIC(10,2) CHECK (height IS NULL OR height > 0),
    shape VARCHAR(100),
    color VARCHAR(100),
    clarity VARCHAR(100),
    origin VARCHAR(150),
    treatment VARCHAR(150),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
