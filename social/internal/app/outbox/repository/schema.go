package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const tableOutboxEvents = "public.outbox_events"

const (
	columnOutboxID            = "id"
	columnOutboxAggType       = "aggregate_type"
	columnOutboxAggID         = "aggregate_id"
	columnOutboxEventType     = "event_type"
	columnOutboxPayload       = "payload"
	columnOutboxCreatedAt     = "created_at"
	columnOutboxPublishedAt   = "published_at"
	columnOutboxRetryCount    = "retry_count"
	columnOutboxNextAttemptAt = "next_attempt_at"
)

type outboxEvent struct {
	ID            uuid.UUID           `db:"id"`
	AggregateType string              `db:"aggregate_type"`
	AggregateID   string              `db:"aggregate_id"`
	EventType     string              `db:"event_type"`
	Payload       []byte              `db:"payload"` // JSONB
	CreatedAt     time.Time           `db:"created_at"`
	PublishedAt   sql.Null[time.Time] `db:"published_at"`
	RetryCount    int                 `db:"retry_count"`
	NextAttemptAt sql.Null[time.Time] `db:"next_attempt_at"`
}

var tableOutboxEventsColumns = []string{
	columnOutboxID,
	columnOutboxAggType,
	columnOutboxAggID,
	columnOutboxEventType,
	columnOutboxPayload,
	columnOutboxCreatedAt,
	columnOutboxPublishedAt,
	columnOutboxRetryCount,
	columnOutboxNextAttemptAt,
}

func (e *outboxEvent) mapFields() map[string]any {
	return map[string]any{
		columnOutboxID:            e.ID,
		columnOutboxAggType:       e.AggregateType,
		columnOutboxAggID:         e.AggregateID,
		columnOutboxEventType:     e.EventType,
		columnOutboxPayload:       e.Payload,
		columnOutboxCreatedAt:     e.CreatedAt,
		columnOutboxPublishedAt:   e.PublishedAt,
		columnOutboxRetryCount:    e.RetryCount,
		columnOutboxNextAttemptAt: e.NextAttemptAt,
	}
}

func (e *outboxEvent) Values(columns ...string) []any {
	m := e.mapFields()
	values := make([]any, 0, len(columns))
	for i := range columns {
		if v, ok := m[columns[i]]; ok {
			values = append(values, v)
		} else {
			values = append(values, nil)
		}
	}
	return values
}
