CREATE TABLE public.email_deliveries (
    email_delivery_id BIGSERIAL PRIMARY KEY,
    notification_id BIGINT NOT NULL
        REFERENCES public.notifications(notification_id)
        ON DELETE CASCADE,
    recipient_email VARCHAR(255) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',
    provider_message_id VARCHAR(255),
    sent_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    error_message TEXT
);
