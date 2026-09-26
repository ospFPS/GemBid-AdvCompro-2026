CREATE TABLE public.subscriptions (
    subscription_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES public.users(user_id)
        ON DELETE CASCADE,
    plan VARCHAR(20) NOT NULL DEFAULT 'FREE'
        CHECK (plan IN ('FREE','PRO')),
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_date IS NULL OR end_date >= start_date)
);
