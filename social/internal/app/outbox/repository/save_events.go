package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"lib/postgres"

	appoutbox "social/internal/app/outbox/processor"
)

func (r *Repository) SaveEvent(ctx context.Context, e *appoutbox.Event) error {
	const api = "outbox.Repository.SaveEvents"
	log.Println(api, "saving event", e.ID)

	row := outboxEvent{
		ID:            e.ID,
		AggregateType: string(e.AggregateType),
		AggregateID:   e.AggregateID,
		EventType:     string(e.EventType),
		Payload:       notnullJSON(e.Payload),
		CreatedAt:     e.CreatedAt,
		PublishedAt: func(t *time.Time) sql.Null[time.Time] {
			if e.PublishedAt != nil {
				return sql.Null[time.Time]{V: *e.PublishedAt, Valid: true}
			}
			return sql.Null[time.Time]{}
		}(&e.CreatedAt),
		RetryCount: e.RetryCount,
	}

	qb := r.qb.Insert(tableOutboxEvents).
		Columns(tableOutboxEventsColumns...).
		Values(row.Values(tableOutboxEventsColumns...)...)

	conn := r.db.GetQueryEngine(ctx)
	if _, err := conn.Execx(ctx, qb); err != nil {
		log.Println(api, "error saving event", e.ID, err)
		return fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	log.Println(api, "successfully saved event", e.ID)
	return nil
}

func notnullJSON(data []byte) []byte {
	if data == nil {
		return []byte("[]")
	}
	return data
}
