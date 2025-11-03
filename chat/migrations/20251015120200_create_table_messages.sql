-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.messages (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id    TEXT NOT NULL,
    owner_id   TEXT NOT NULL,
    text       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_id_created_at ON public.messages(chat_id, created_at);

COMMENT ON TABLE public.messages IS 'Таблица сообщений в чатах';

COMMENT ON COLUMN public.messages.id         IS 'Уникальный идентификатор сообщения';
COMMENT ON COLUMN public.messages.chat_id    IS 'Идентификатор чата';
COMMENT ON COLUMN public.messages.owner_id   IS 'Идентификатор владельца/отправителя сообщения';
COMMENT ON COLUMN public.messages.text       IS 'Текст сообщения';
COMMENT ON COLUMN public.messages.created_at IS 'Дата и время отправки сообщения';
COMMENT ON COLUMN public.messages.updated_at IS 'Дата и время последнего обновления сообщения (nullable)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS public.idx_messages_chat_id_created_at;
DROP TABLE IF EXISTS public.messages;
-- +goose StatementEnd