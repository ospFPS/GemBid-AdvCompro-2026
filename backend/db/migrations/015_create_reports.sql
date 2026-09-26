CREATE TABLE public.reports (
    report_id BIGSERIAL PRIMARY KEY,
    reporter_user_id BIGINT NOT NULL
        REFERENCES public.users(user_id)
        ON DELETE RESTRICT,
    reported_user_id BIGINT
        REFERENCES public.users(user_id)
        ON DELETE SET NULL,
    listing_id BIGINT
        REFERENCES public.listings(listing_id)
        ON DELETE SET NULL,
    auction_id BIGINT
        REFERENCES public.auctions(auction_id)
        ON DELETE SET NULL,
    reason VARCHAR(150) NOT NULL,
    description TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    reviewed_by BIGINT
        REFERENCES public.users(user_id)
        ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);
