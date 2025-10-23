-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.outbox_events (
    id              UUID        NOT NULL,
    aggregate_type  TEXT        NOT NULL,
    aggregate_id    TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    payload         JSONB       NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at    TIMESTAMPTZ,
    retry_count     INT         NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ
    ) PARTITION BY RANGE (created_at);

COMMENT ON TABLE public.outbox_events IS 'Таблица событий для паттерна Transactional Outbox (partitioned)';
COMMENT ON COLUMN public.outbox_events.id               IS 'Уникальный идентификатор события (UUID)';
COMMENT ON COLUMN public.outbox_events.aggregate_type   IS 'Тип агрегата (например, order, user)';
COMMENT ON COLUMN public.outbox_events.aggregate_id     IS 'Идентификатор агрегата (например, order_id)';
COMMENT ON COLUMN public.outbox_events.event_type       IS 'Тип события (например, OrderCreated)';
COMMENT ON COLUMN public.outbox_events.payload          IS 'Тело события JSONB';
COMMENT ON COLUMN public.outbox_events.created_at       IS 'Время записи события в outbox';
COMMENT ON COLUMN public.outbox_events.published_at     IS 'Время успешной публикации события';
COMMENT ON COLUMN public.outbox_events.retry_count      IS 'Количество попыток публикации';
COMMENT ON COLUMN public.outbox_events.next_attempt_at  IS 'Время следующей попытки публикации (NULL = сразу)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.outbox_events;
-- +goose StatementEnd