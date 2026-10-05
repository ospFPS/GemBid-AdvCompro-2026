BEGIN;

ALTER TABLE public.users
    ADD COLUMN username VARCHAR(30);

UPDATE public.users
SET username = 'user_' || user_id
WHERE username IS NULL OR BTRIM(username) = '';

ALTER TABLE public.users
    ALTER COLUMN username SET NOT NULL,
    ALTER COLUMN email DROP NOT NULL;

CREATE UNIQUE INDEX uq_users_username_lower
    ON public.users (LOWER(username));

ALTER TABLE public.users
    ADD CONSTRAINT users_username_format_check
        CHECK (username ~ '^[a-z0-9._-]{3,30}$');

COMMIT;

