CREATE TABLE public.auctions (
    auction_id BIGSERIAL PRIMARY KEY,
    listing_id BIGINT NOT NULL UNIQUE
        REFERENCES public.listings(listing_id)
        ON DELETE CASCADE,
    starting_price NUMERIC(12,2) NOT NULL CHECK (starting_price >= 0),
    minimum_increment NUMERIC(12,2) NOT NULL CHECK (minimum_increment > 0),
    reserve_price NUMERIC(12,2) CHECK (reserve_price IS NULL OR reserve_price >= 0),
    current_price NUMERIC(12,2) NOT NULL CHECK (current_price >= 0),
    current_highest_bidder_id BIGINT
        REFERENCES public.users(user_id)
        ON DELETE SET NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED',
    winner_id BIGINT
        REFERENCES public.users(user_id)
        ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    CHECK (end_time > start_time),
    CHECK (reserve_price IS NULL OR reserve_price >= starting_price),
    CHECK (current_price >= starting_price)
);
