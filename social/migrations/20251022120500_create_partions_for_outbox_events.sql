-- +goose Up
-- +goose StatementBegin

-- Месячная партиция (range) OCT-2025 → сама partitioned BY LIST (aggregate_type)
CREATE TABLE IF NOT EXISTS public.outbox_events_2025_10
    PARTITION OF public.outbox_events
    FOR VALUES FROM ('2025-10-01 00:00:00+00') TO ('2025-11-01 00:00:00+00')
    PARTITION BY LIST (aggregate_type);

COMMENT ON TABLE public.outbox_events_2025_10 IS 'Партиция outbox за октябрь 2025';

-- Партиция по aggregate_type='friend_request' (и внутри — LIST по event_type)
CREATE TABLE IF NOT EXISTS public.outbox_events_2025_10__agg_friend_request
    PARTITION OF public.outbox_events_2025_10
    FOR VALUES IN ('friend_request')
    PARTITION BY LIST (event_type);

COMMENT ON TABLE public.outbox_events_2025_10__agg_friend_request IS 'Партиция для aggregate_type=friend_request (внутри — по event_type)';

-- Лист для event_type='FriendRequestCreated'
CREATE TABLE IF NOT EXISTS public.outbox_events_2025_10__agg_friend_request__evt_created
    PARTITION OF public.outbox_events_2025_10__agg_friend_request
    FOR VALUES IN ('FriendRequestCreated');

COMMENT ON TABLE public.outbox_events_2025_10__agg_friend_request__evt_created IS 'Лист: friend_request / FriendRequestCreated';

-- Индексы для FriendRequestCreated
CREATE INDEX IF NOT EXISTS idx_o_2025_10_fr_created_due
    ON public.outbox_events_2025_10__agg_friend_request__evt_created (next_attempt_at)
    WHERE published_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_o_2025_10_fr_created_unpub
    ON public.outbox_events_2025_10__agg_friend_request__evt_created (published_at)
    WHERE published_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_o_2025_10_fr_created_ordering
    ON public.outbox_events_2025_10__agg_friend_request__evt_created (created_at, id);

-- Лист для event_type='FriendRequestStatusUpdated'
CREATE TABLE IF NOT EXISTS public.outbox_events_2025_10__agg_friend_request__evt_status_updated
    PARTITION OF public.outbox_events_2025_10__agg_friend_request
    FOR VALUES IN ('FriendRequestStatusUpdated');

COMMENT ON TABLE public.outbox_events_2025_10__agg_friend_request__evt_status_updated IS 'Лист: friend_request / FriendRequestStatusUpdated';

-- Индексы для FriendRequestStatusUpdated
CREATE INDEX IF NOT EXISTS idx_o_2025_10_fr_status_updated_due
    ON public.outbox_events_2025_10__agg_friend_request__evt_status_updated (next_attempt_at)
    WHERE published_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_o_2025_10_fr_status_updated_unpub
    ON public.outbox_events_2025_10__agg_friend_request__evt_status_updated (published_at)
    WHERE published_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_o_2025_10_fr_status_updated_ordering
    ON public.outbox_events_2025_10__agg_friend_request__evt_status_updated (created_at, id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd