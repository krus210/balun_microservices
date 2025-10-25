package repository

import (
	"database/sql"
	"time"

	"notifications/internal/app/models"
)

// ToModel конвертирует inboxMessage (схема БД) в доменную модель models.InboxMessage
func ToModel(r *inboxMessage) *models.InboxMessage {
	if r == nil {
		return nil
	}

	var lastError *string
	if r.LastError.Valid {
		lastError = &r.LastError.V
	}

	var processedAt *time.Time
	if r.ProcessedAt.Valid {
		processedAt = &r.ProcessedAt.V
	}

	return &models.InboxMessage{
		ID:          r.ID,
		Topic:       r.Topic,
		Partition:   r.Partition,
		Offset:      r.Offset,
		Payload:     r.Payload,
		Status:      r.Status,
		Attempts:    r.Attempts,
		LastError:   lastError,
		ReceivedAt:  r.ReceivedAt,
		ProcessedAt: processedAt,
	}
}

// FromModel конвертирует доменную модель в inboxMessage (для INSERT/UPDATE)
func FromModel(m *models.InboxMessage) inboxMessage {
	if m == nil {
		return inboxMessage{}
	}

	var lastError sql.Null[string]
	if m.LastError != nil {
		lastError = sql.Null[string]{V: *m.LastError, Valid: true}
	}

	var processedAt sql.Null[time.Time]
	if m.ProcessedAt != nil {
		processedAt = sql.Null[time.Time]{V: *m.ProcessedAt, Valid: true}
	}

	return inboxMessage{
		ID:          m.ID,
		Topic:       m.Topic,
		Partition:   m.Partition,
		Offset:      m.Offset,
		Payload:     m.Payload,
		Status:      m.Status,
		Attempts:    m.Attempts,
		LastError:   lastError,
		ReceivedAt:  m.ReceivedAt,
		ProcessedAt: processedAt,
	}
}
