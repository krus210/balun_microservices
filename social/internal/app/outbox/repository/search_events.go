package repository

import (
	"context"
	"time"

	appoutbox "social/internal/app/outbox/processor"

	"github.com/Masterminds/squirrel"
)

// SearchEvents выбирает события из outbox по заданным опциям.
// Возвращает пустой срез при ошибке (сигнатура без error).
func (r *Repository) SearchEvents(ctx context.Context, opts ...appoutbox.SearchEventsOption) []*appoutbox.Event {
	o := appoutbox.CollectSearchEventsOptions(opts...)

	// Базовый селект
	qb := r.qb.
		Select(tableOutboxEventsColumns...).
		From(tableOutboxEvents).
		OrderBy(columnOutboxCreatedAt).
		Limit(uint64(o.Limit))

	// Фильтры
	if o.OnlyUnpublished {
		qb = qb.Where(squirrel.Eq{columnOutboxPublishedAt: nil}) // IS NULL
	}
	// retry_count <= MaxRetryCount
	qb = qb.Where(squirrel.LtOrEq{columnOutboxRetryCount: o.MaxRetryCount})

	if o.AggregateType != nil {
		qb = qb.Where(squirrel.Eq{columnOutboxAggType: string(*o.AggregateType)})
	}
	if o.EventType != nil {
		qb = qb.Where(squirrel.Eq{columnOutboxEventType: string(*o.EventType)})
	}
	if o.NotBefore != nil {
		qb = qb.Where(squirrel.GtOrEq{columnOutboxCreatedAt: *o.NotBefore})
	}
	if o.NotAfter != nil {
		qb = qb.Where(squirrel.LtOrEq{columnOutboxCreatedAt: *o.NotAfter})
	}
	if o.DueAt != nil {
		// next_attempt_at IS NULL OR next_attempt_at <= dueAt
		qb = qb.Where(
			squirrel.Or{
				squirrel.Eq{columnOutboxNextAttemptAt: nil},
				squirrel.LtOrEq{columnOutboxNextAttemptAt: *o.DueAt},
			},
		)
	}

	// Блокировка строк для конкурентных воркеров
	if o.WithLock {
		qb = qb.Suffix("FOR UPDATE SKIP LOCKED")
	}

	// Выполнение
	conn := r.db.GetQueryEngine(ctx)
	var rows []outboxEvent
	if err := conn.Selectx(ctx, &rows, qb); err != nil {
		return nil
	}

	// Маппинг в доменную модель
	events := make([]*appoutbox.Event, 0, len(rows))
	for i := range rows {
		var publishedAt *time.Time
		if rows[i].PublishedAt.Valid {
			t := rows[i].PublishedAt.V
			publishedAt = &t
		}
		events = append(events, &appoutbox.Event{
			ID:            rows[i].ID,
			AggregateType: appoutbox.AggregateType(rows[i].AggregateType),
			AggregateID:   rows[i].AggregateID,
			EventType:     appoutbox.EventType(rows[i].EventType),
			Payload:       rows[i].Payload,
			CreatedAt:     rows[i].CreatedAt,
			PublishedAt:   publishedAt,
			RetryCount:    rows[i].RetryCount,
		})
	}
	return events
}
