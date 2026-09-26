CREATE TABLE public.orders (
    order_id BIGSERIAL PRIMARY KEY,
    listing_id BIGINT NOT NULL
        REFERENCES public.listings(listing_id)
        ON DELETE RESTRICT,
    buyer_id BIGINT NOT NULL
        REFERENCES public.users(user_id)
        ON DELETE RESTRICT,
    seller_id BIGINT NOT NULL
        REFERENCES public.seller_profiles(seller_id)
        ON DELETE RESTRICT,
    auction_id BIGINT
        REFERENCES public.auctions(auction_id)
        ON DELETE SET NULL,
    winning_bid_id BIGINT
        REFERENCES public.bids(bid_id)
        ON DELETE SET NULL,
    item_price NUMERIC(12,2) NOT NULL CHECK (item_price >= 0),
    shipping_fee NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (shipping_fee >= 0),
    total_amount NUMERIC(12,2) NOT NULL CHECK (total_amount >= 0),
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
