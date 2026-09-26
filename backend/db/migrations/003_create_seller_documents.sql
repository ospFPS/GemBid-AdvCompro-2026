CREATE TABLE public.seller_documents (
    document_id BIGSERIAL PRIMARY KEY,
    seller_id BIGINT NOT NULL
        REFERENCES public.seller_profiles(seller_id)
        ON DELETE CASCADE,
    document_type VARCHAR(100) NOT NULL,
    document_url TEXT NOT NULL,
    verification_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
