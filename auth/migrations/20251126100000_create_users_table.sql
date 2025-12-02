-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON public.users(email);

COMMENT ON TABLE public.users IS 'Таблица пользователей с хешированными паролями';
COMMENT ON COLUMN public.users.id IS 'UUID пользователя';
COMMENT ON COLUMN public.users.email IS 'Email пользователя (уникальный)';
COMMENT ON COLUMN public.users.password_hash IS 'Bcrypt хеш пароля';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.users;
-- +goose StatementEnd
