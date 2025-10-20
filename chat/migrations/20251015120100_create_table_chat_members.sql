-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.chat_members (
    chat_id BIGINT NOT NULL REFERENCES public.chats(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    UNIQUE(chat_id, user_id)
);

COMMENT ON TABLE public.chat_members IS 'Таблица участников чатов';

COMMENT ON COLUMN public.chat_members.chat_id IS 'Идентификатор чата';
COMMENT ON COLUMN public.chat_members.user_id IS 'Идентификатор пользователя';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.chat_members;
-- +goose StatementEnd