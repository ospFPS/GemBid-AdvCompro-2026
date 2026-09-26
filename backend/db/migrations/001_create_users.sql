CREATE TABLE users (
    user_id BIGSERIAL PRIMARY KEY,

    full_name VARCHAR(150) NOT NULL,

    email VARCHAR(255) NOT NULL UNIQUE,

    password_hash TEXT NOT NULL,

    phone VARCHAR(30),

    role VARCHAR(20) NOT NULL DEFAULT 'USER'
        CHECK (role IN ('USER', 'ADMIN')),

    seller_status VARCHAR(20) NOT NULL DEFAULT 'NONE'
        CHECK (
            seller_status IN (
                'NONE',
                'PENDING',
                'VERIFIED',
                'REJECTED',
                'SUSPENDED'
            )
        ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);