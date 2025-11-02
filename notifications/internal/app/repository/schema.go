package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const InboxMessagesTable = "public.inbox_messages"

const (
	InboxMessagesTableColumnID          = "id"
	InboxMessagesTableColumnTopic       = "topic"
	InboxMessagesTableColumnPartition   = "partition"
	InboxMessagesTableColumnOffset      = "kafka_offset"
	InboxMessagesTableColumnPayload     = "payload"
	InboxMessagesTableColumnStatus      = "status"
	InboxMessagesTableColumnAttempts    = "attempts"
	InboxMessagesTableColumnLastError   = "last_error"
	InboxMessagesTableColumnReceivedAt  = "received_at"
	InboxMessagesTableColumnProcessedAt = "processed_at"
)

const (
	InboxMessageStatusReceived   = "received"
	InboxMessageStatusProcessing = "processing"
	InboxMessageStatusFailed     = "failed"
	InboxMessageStatusProcessed  = "processed"
)

type inboxMessage struct {
	ID          uuid.UUID           `db:"id"`
	Topic       string              `db:"topic"`
	Partition   int                 `db:"partition"`
	Offset      int64               `db:"kafka_offset"`
	Payload     []byte              `db:"payload"` // JSONB
	Status      string              `db:"status"`
	Attempts    int                 `db:"attempts"`
	LastError   sql.Null[string]    `db:"last_error"`
	ReceivedAt  time.Time           `db:"received_at"`
	ProcessedAt sql.Null[time.Time] `db:"processed_at"`
}

var InboxMessagesTableColumns = []string{
	InboxMessagesTableColumnID,
	InboxMessagesTableColumnTopic,
	InboxMessagesTableColumnPartition,
	InboxMessagesTableColumnOffset,
	InboxMessagesTableColumnPayload,
	InboxMessagesTableColumnStatus,
	InboxMessagesTableColumnAttempts,
	InboxMessagesTableColumnLastError,
	InboxMessagesTableColumnReceivedAt,
	InboxMessagesTableColumnProcessedAt,
}

func (m *inboxMessage) mapFields() map[string]any {
	return map[string]any{
		InboxMessagesTableColumnID:          m.ID,
		InboxMessagesTableColumnTopic:       m.Topic,
		InboxMessagesTableColumnPartition:   m.Partition,
		InboxMessagesTableColumnOffset:      m.Offset,
		InboxMessagesTableColumnPayload:     m.Payload,
		InboxMessagesTableColumnStatus:      m.Status,
		InboxMessagesTableColumnAttempts:    m.Attempts,
		InboxMessagesTableColumnLastError:   m.LastError,
		InboxMessagesTableColumnReceivedAt:  m.ReceivedAt,
		InboxMessagesTableColumnProcessedAt: m.ProcessedAt,
	}
}

func (m *inboxMessage) Values(columns ...string) []any {
	fieldMap := m.mapFields()
	values := make([]any, 0, len(columns))
	for i := range columns {
		if v, ok := fieldMap[columns[i]]; ok {
			values = append(values, v)
		} else {
			values = append(values, nil)
		}
	}
	return values
}
