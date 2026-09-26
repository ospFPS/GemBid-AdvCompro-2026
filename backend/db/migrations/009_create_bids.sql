CREATE TABLE public.bids (
    bid_id BIGSERIAL PRIMARY KEY,
    auction_id BIGINT NOT NULL
        REFERENCES public.auctions(auction_id)
        ON DELETE RESTRICT,
    bidder_id BIGINT NOT NULL
        REFERENCES public.users(user_id)
        ON DELETE RESTRICT,
    bid_amount NUMERIC(12,2) NOT NULL CHECK (bid_amount > 0),
    bid_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    bid_status VARCHAR(20) NOT NULL DEFAULT 'ACCEPTED'
);

ALTER TABLE public.auctions
ADD COLUMN winning_bid_id BIGINT
    REFERENCES public.bids(bid_id)
    ON DELETE SET NULL;
