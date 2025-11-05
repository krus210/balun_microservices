-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.friend_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id TEXT NOT NULL,
    to_user_id TEXT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

-- Индексы для быстрого поиска заявок
CREATE INDEX idx_friend_requests_from_user_id ON public.friend_requests(from_user_id);
CREATE INDEX idx_friend_requests_to_user_id ON public.friend_requests(to_user_id);
CREATE INDEX idx_friend_requests_status ON public.friend_requests(status);

-- Уникальный индекс для предотвращения дублирования заявок
CREATE UNIQUE INDEX idx_friend_requests_users ON public.friend_requests(from_user_id, to_user_id);

COMMENT ON TABLE public.friend_requests IS 'Таблица заявок в друзья';

COMMENT ON COLUMN public.friend_requests.id IS 'Уникальный идентификатор заявки';
COMMENT ON COLUMN public.friend_requests.from_user_id IS 'ID пользователя, отправившего заявку';
COMMENT ON COLUMN public.friend_requests.to_user_id IS 'ID пользователя, получившего заявку';
COMMENT ON COLUMN public.friend_requests.status IS 'Статус заявки (0 - ожидает, 1 - принята, 2 - отклонена)';
COMMENT ON COLUMN public.friend_requests.created_at IS 'Дата и время создания заявки';
COMMENT ON COLUMN public.friend_requests.updated_at IS 'Дата и время последнего обновления заявки (nullable)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.friend_requests;
-- +goose StatementEnd
