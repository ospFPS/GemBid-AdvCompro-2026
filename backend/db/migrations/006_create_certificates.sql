CREATE TABLE public.certificates (
    certificate_id BIGSERIAL PRIMARY KEY,
    gem_id BIGINT NOT NULL
        REFERENCES public.gems(gem_id)
        ON DELETE CASCADE,
    issuer VARCHAR(150) NOT NULL,
    certificate_number VARCHAR(150),
    certificate_url TEXT,
    issued_date DATE,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'UNVERIFIED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
