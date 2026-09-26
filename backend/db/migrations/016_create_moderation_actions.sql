CREATE TABLE public.moderation_actions (
    action_id BIGSERIAL PRIMARY KEY,
    admin_user_id BIGINT NOT NULL
        REFERENCES public.users(user_id)
        ON DELETE RESTRICT,
    target_user_id BIGINT
        REFERENCES public.users(user_id)
        ON DELETE SET NULL,
    listing_id BIGINT
        REFERENCES public.listings(listing_id)
        ON DELETE SET NULL,
    auction_id BIGINT
        REFERENCES public.auctions(auction_id)
        ON DELETE SET NULL,
    action_type VARCHAR(100) NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
