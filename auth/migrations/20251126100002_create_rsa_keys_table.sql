-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.rsa_keys (
    kid TEXT PRIMARY KEY,
    private_key_pem TEXT NOT NULL,
    public_key_pem TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'next', 'expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_rsa_keys_status ON public.rsa_keys(status);

COMMENT ON TABLE public.rsa_keys IS 'RSA ключи для подписи JWT (fallback для dev, основное хранение в Vault)';
COMMENT ON COLUMN public.rsa_keys.kid IS 'Key ID (уникальный идентификатор ключа)';
COMMENT ON COLUMN public.rsa_keys.status IS 'Статус ключа: active (основной), next (следующий), expired (истекший)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.rsa_keys;
-- +goose StatementEnd
