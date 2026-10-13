BEGIN;
UPDATE public.orders SET status = 'PENDING_PAYMENT' WHERE status = 'PENDING';

ALTER TABLE public.orders
    ALTER COLUMN status SET DEFAULT 'PENDING_PAYMENT',
    ADD CONSTRAINT orders_status_check
        CHECK (status IN ('PENDING_PAYMENT','PAID','SHIPPED','DELIVERED','CANCELLED'));

CREATE UNIQUE INDEX uq_orders_one_open_per_listing
    ON public.orders(listing_id)
    WHERE status <> 'CANCELLED';

ALTER TABLE public.transactions
    ALTER COLUMN order_id DROP NOT NULL,
    ADD COLUMN subscription_id BIGINT
        REFERENCES public.subscriptions(subscription_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT transactions_target_check
        CHECK (num_nonnulls(order_id, subscription_id) = 1),
    ADD CONSTRAINT transactions_status_check
        CHECK (status IN ('PENDING','SUCCESS','FAILED','REFUNDED'));

CREATE INDEX idx_transactions_subscription_id
    ON public.transactions(subscription_id);

CREATE INDEX idx_transactions_status_paid_at
    ON public.transactions(status, paid_at);

ALTER TABLE public.shipping
    ADD CONSTRAINT shipping_status_check
        CHECK (status IN ('PENDING','SHIPPED','DELIVERED'));

ALTER TABLE public.email_deliveries
    ADD CONSTRAINT email_deliveries_status_check
        CHECK (status IN ('PENDING','SENT','FAILED'));

CREATE INDEX idx_email_deliveries_pending
    ON public.email_deliveries(email_delivery_id)
    WHERE status = 'PENDING';

ALTER TABLE public.subscriptions
    ADD CONSTRAINT subscriptions_status_check
        CHECK (status IN ('ACTIVE','CANCELLED','EXPIRED'));

CREATE UNIQUE INDEX uq_subscriptions_one_active_per_user
    ON public.subscriptions(user_id)
    WHERE status = 'ACTIVE';

COMMIT;
