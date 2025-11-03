-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS public.chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);

COMMENT ON TABLE public.chats IS 'Таблица чатов';

COMMENT ON COLUMN public.chats.id         IS 'Уникальный идентификатор чата';
COMMENT ON COLUMN public.chats.created_at IS 'Дата и время создания чата';
COMMENT ON COLUMN public.chats.updated_at IS 'Дата и время последнего обновления чата (nullable)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.chats;
-- +goose StatementEnd