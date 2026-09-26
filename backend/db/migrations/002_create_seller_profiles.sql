CREATE TABLE public.seller_profiles (
    seller_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE
        REFERENCES public.users(user_id)
        ON DELETE CASCADE,
    shop_name VARCHAR(150) NOT NULL,
    seller_type VARCHAR(50),
    legal_name VARCHAR(150) NOT NULL,
    address TEXT,
    description TEXT,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (verification_status IN ('PENDING','VERIFIED','REJECTED','SUSPENDED')),
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ,
    reviewed_by BIGINT
        REFERENCES public.users(user_id)
        ON DELETE SET NULL
);
