-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    jti TEXT NOT NULL UNIQUE,
    device_id TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    replaced_by_jti TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_jti ON public.refresh_tokens(jti);
CREATE INDEX idx_refresh_tokens_token_hash ON public.refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON public.refresh_tokens(expires_at) WHERE used_at IS NULL;

COMMENT ON TABLE public.refresh_tokens IS 'Таблица refresh токенов с защитой от переиспользования';
COMMENT ON COLUMN public.refresh_tokens.token_hash IS 'SHA-256 хеш refresh токена';
COMMENT ON COLUMN public.refresh_tokens.jti IS 'JWT ID (уникальный идентификатор токена)';
COMMENT ON COLUMN public.refresh_tokens.device_id IS 'Опциональный ID устройства для управления сессиями';
COMMENT ON COLUMN public.refresh_tokens.used_at IS 'Время использования токена (для rotation)';
COMMENT ON COLUMN public.refresh_tokens.replaced_by_jti IS 'JTI нового токена, которым заменили этот';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.refresh_tokens;
-- +goose StatementEnd
