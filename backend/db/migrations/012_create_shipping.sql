CREATE TABLE public.shipping (
    shipping_id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE
        REFERENCES public.orders(order_id)
        ON DELETE CASCADE,
    courier VARCHAR(100),
    tracking_number VARCHAR(150),
    shipping_address TEXT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    shipped_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
