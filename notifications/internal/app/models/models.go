package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	InboxMessageStatusReceived   = "received"
	InboxMessageStatusProcessing = "processing"
	InboxMessageStatusFailed     = "failed"
	InboxMessageStatusProcessed  = "processed"
)

// InboxMessage представляет входящее сообщение из Kafka для обработки
type InboxMessage struct {
	ID          uuid.UUID
	Topic       string
	Partition   int
	Offset      int64
	Payload     []byte
	Status      string
	Attempts    int
	LastError   *string
	ReceivedAt  time.Time
	ProcessedAt *time.Time
}
