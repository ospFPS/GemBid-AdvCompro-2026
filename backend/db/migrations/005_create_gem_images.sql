CREATE TABLE public.gem_images (
    image_id BIGSERIAL PRIMARY KEY,
    gem_id BIGINT NOT NULL
        REFERENCES public.gems(gem_id)
        ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0 CHECK (display_order >= 0),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
