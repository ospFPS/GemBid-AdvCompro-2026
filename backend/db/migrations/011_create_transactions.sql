CREATE TABLE public.transactions (
    transaction_id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL
        REFERENCES public.orders(order_id)
        ON DELETE RESTRICT,
    buyer_id BIGINT NOT NULL
        REFERENCES public.users(user_id)
        ON DELETE RESTRICT,
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    payment_method VARCHAR(50),
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    reference_number VARCHAR(150),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ
);
