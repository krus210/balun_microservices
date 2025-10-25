package delivery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"notifications/internal/app/models"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

const headerEventID = "event_id"

// Repository интерфейс для работы с хранилищем inbox сообщений
type Repository interface {
	SaveInboxMessage(ctx context.Context, msg *models.InboxMessage) error
}

// InboxHandler обработчик для сохранения входящих событий из Kafka в inbox
type InboxHandler struct {
	repository Repository
}

// NewInboxHandler конструктор InboxHandler
func NewInboxHandler(repository Repository) *InboxHandler {
	return &InboxHandler{
		repository: repository,
	}
}

// SaveInboxMessage преобразует Kafka сообщение в InboxMessage и сохраняет в БД
func (h *InboxHandler) SaveInboxMessage(ctx context.Context, message *sarama.ConsumerMessage) (needMark bool, err error) {
	id, ok := extractID(message)
	if !ok {
		// без ID не можем гарантировать идемпотентность — безопаснее скипнуть и скоммитить,
		// либо отправить в DLQ.
		err = fmt.Errorf("skip message without ID (commit offset): topic=%s p=%d off=%d",
			message.Topic, message.Partition, message.Offset)
		needMark = true
		return needMark, err
	}

	receivedAt := message.Timestamp
	if receivedAt.IsZero() {
		receivedAt = time.Now()
	}

	messageID, err := uuid.Parse(id)
	if err != nil {
		err = fmt.Errorf("invalid UUID in event_id header: %s, err: %v", id, err)
		needMark = true // скипаем невалидный UUID, коммитим offset
		return needMark, err
	}

	inboxMsg := &models.InboxMessage{
		ID:          messageID,
		Topic:       message.Topic,
		Partition:   int(message.Partition),
		Offset:      message.Offset,
		Payload:     message.Value,
		Status:      models.InboxMessageStatusReceived,
		Attempts:    0,
		LastError:   nil,
		ReceivedAt:  receivedAt,
		ProcessedAt: nil,
	}

	// бизнес-обработка (идемпотентная)
	err = h.repository.SaveInboxMessage(ctx, inboxMsg)
	if err != nil {
		// обработка упала: НЕ коммитим offset -> Kafka переотправит (at-least-once)
		err = fmt.Errorf("handle failed (will retry): id=%s topic=%s p=%d off=%d err=%v",
			id, message.Topic, message.Partition, message.Offset, err)
		needMark = false
		return false, err
	}

	// успех: коммитим offset
	err = nil
	needMark = true
	return needMark, err
}

func extractID(msg *sarama.ConsumerMessage) (string, bool) {
	for _, h := range msg.Headers {
		if strings.EqualFold(string(h.Key), headerEventID) && len(h.Value) > 0 {
			return string(h.Value), true
		}
	}
	return "", false
}
