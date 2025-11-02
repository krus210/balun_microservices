-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.inbox_messages (
    id           UUID        PRIMARY KEY,
    topic        TEXT        NOT NULL,
    partition    INT         NOT NULL,
    kafka_offset BIGINT      NOT NULL,
    payload      JSONB       NOT NULL,
    status       TEXT        NOT NULL,
    attempts     INT         NOT NULL DEFAULT 0,
    last_error   TEXT,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);

COMMENT ON TABLE public.inbox_messages IS 'Таблица входящих сообщений из Kafka для паттерна Transactional Inbox';
COMMENT ON COLUMN public.inbox_messages.id           IS 'Уникальный идентификатор сообщения (event_id из продьюсера)';
COMMENT ON COLUMN public.inbox_messages.topic        IS 'Kafka топик, из которого получено сообщение';
COMMENT ON COLUMN public.inbox_messages.partition    IS 'Номер партиции Kafka';
COMMENT ON COLUMN public.inbox_messages.kafka_offset IS 'Offset сообщения в партиции Kafka';
COMMENT ON COLUMN public.inbox_messages.payload      IS 'Тело сообщения в формате JSONB';
COMMENT ON COLUMN public.inbox_messages.status       IS 'Статус обработки: received | processing | processed | failed';
COMMENT ON COLUMN public.inbox_messages.attempts     IS 'Количество попыток обработки';
COMMENT ON COLUMN public.inbox_messages.last_error   IS 'Текст последней ошибки при обработке';
COMMENT ON COLUMN public.inbox_messages.received_at  IS 'Время получения сообщения из Kafka';
COMMENT ON COLUMN public.inbox_messages.processed_at IS 'Время успешной обработки сообщения';

-- Индекс для эффективной обработки сообщений по статусу
CREATE INDEX IF NOT EXISTS idx_inbox_messages_status ON public.inbox_messages(status) WHERE status IN ('received', 'processing', 'failed');

-- Уникальный индекс для предотвращения дублирования сообщений из Kafka
CREATE UNIQUE INDEX IF NOT EXISTS idx_inbox_messages_kafka_unique ON public.inbox_messages(topic, partition, kafka_offset);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.inbox_messages;
-- +goose StatementEnd