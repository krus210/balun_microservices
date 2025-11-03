-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.user_profiles (
    id UUID NOT NULL PRIMARY KEY,
    nickname VARCHAR(255) NOT NULL,
    bio TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

-- Уникальные индексы для nickname
CREATE UNIQUE INDEX idx_user_profiles_nickname ON public.user_profiles(nickname);

-- Индекс для поиска по created_at
CREATE INDEX idx_user_profiles_created_at ON public.user_profiles(created_at);

COMMENT ON TABLE public.user_profiles IS 'Таблица профилей пользователей';

COMMENT ON COLUMN public.user_profiles.id IS 'Уникальный идентификатор пользователя UUID (передается из auth сервиса)';
COMMENT ON COLUMN public.user_profiles.nickname IS 'Никнейм пользователя (уникальный)';
COMMENT ON COLUMN public.user_profiles.bio IS 'Биография пользователя (nullable)';
COMMENT ON COLUMN public.user_profiles.avatar_url IS 'URL аватара пользователя (nullable)';
COMMENT ON COLUMN public.user_profiles.created_at IS 'Дата и время создания профиля';
COMMENT ON COLUMN public.user_profiles.updated_at IS 'Дата и время последнего обновления профиля (nullable)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.user_profiles;
-- +goose StatementEnd
