CREATE TABLE public.notifications (
    notification_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES public.users(user_id)
        ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    auction_id BIGINT
        REFERENCES public.auctions(auction_id)
        ON DELETE SET NULL,
    listing_id BIGINT
        REFERENCES public.listings(listing_id)
        ON DELETE SET NULL,
    order_id BIGINT
        REFERENCES public.orders(order_id)
        ON DELETE SET NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
