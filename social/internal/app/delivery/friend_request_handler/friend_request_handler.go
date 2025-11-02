package friend_request_handler

import (
	"context"
	"errors"
	"log"
	"slices"
	"time"

	outbox "social/internal/app/outbox/processor"

	"github.com/IBM/sarama" // Shopify/sarama
	"github.com/google/uuid"
)

type TopicResolver func(e *outbox.Event) (topic string, key string)

// KafkaFriendRequestBatchHandler реализует FriendRequestBatchHandler.
// Использует sarama.SyncProducer и отправляет батчами.
type KafkaFriendRequestBatchHandler struct {
	producer     sarama.SyncProducer
	resolve      TopicResolver
	maxBatchSize int
	closeTimeout time.Duration
}

// ===== Опции =====

type KafkaHandlerOption func(*KafkaFriendRequestBatchHandler)

// WithTopic фиксирует один топик, ключ = AggregateID.
func WithTopic(topic string) KafkaHandlerOption {
	return func(h *KafkaFriendRequestBatchHandler) {
		h.resolve = func(e *outbox.Event) (string, string) { return topic, e.AggregateID }
	}
}

// WithTopicResolver позволяет выбрать топик/ключ динамически.
func WithTopicResolver(r TopicResolver) KafkaHandlerOption {
	return func(h *KafkaFriendRequestBatchHandler) { h.resolve = r }
}

// WithMaxBatchSize настраивает размер чанка для SendMessages (по умолчанию 500).
func WithMaxBatchSize(n int) KafkaHandlerOption {
	return func(h *KafkaFriendRequestBatchHandler) {
		if n > 0 {
			h.maxBatchSize = n
		}
	}
}

// WithCloseTimeout задаёт таймаут при закрытии продюсера.
func WithCloseTimeout(d time.Duration) KafkaHandlerOption {
	return func(h *KafkaFriendRequestBatchHandler) { h.closeTimeout = d }
}

// ===== Конструктор =====

// NewKafkaFriendRequestBatchHandler создаёт идемпотентный sync-producer.
func NewKafkaFriendRequestBatchHandler(producer sarama.SyncProducer, opts ...KafkaHandlerOption) *KafkaFriendRequestBatchHandler {
	h := &KafkaFriendRequestBatchHandler{
		producer:     producer,
		maxBatchSize: 500,
		closeTimeout: 5 * time.Second,
		// по умолчанию — один топик "friend-request-events", ключ = AggregateID
		resolve: func(e *outbox.Event) (string, string) { return "friend-request-events", e.AggregateID },
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Close аккуратно закрывает продюсер.
func (h *KafkaFriendRequestBatchHandler) Close() error {
	done := make(chan struct{})
	var cerr error
	go func() {
		cerr = h.producer.Close()
		close(done)
	}()
	select {
	case <-done:
		return cerr
	case <-time.After(h.closeTimeout):
		return errors.New("kafka: close timeout")
	}
}

const headerEventID = "event_id"

// HandleBatch отправляет события пачками. Возвращает id успешных/ошибочных.
func (h *KafkaFriendRequestBatchHandler) HandleBatch(ctx context.Context, events []*outbox.Event) (succeeded []uuid.UUID, failed []uuid.UUID, err error) {
	if len(events) == 0 {
		log.Println("KafkaFriendRequestBatchHandler", "nothing to send")
		return nil, nil, nil
	}

	defer func() {
		if err != nil {
			log.Println("HandleBatch", err)
		} else {
			log.Println("HandleBatch", "succeeded", succeeded, "failed", failed)
		}
	}()

	chunks := chunkEvents(events, h.maxBatchSize)
	succeeded = make([]uuid.UUID, 0, len(events))
	failed = make([]uuid.UUID, 0, len(events))

	for _, evs := range chunks {
		select {
		case <-ctx.Done():
			return succeeded, append(failed, ids(evs)...), ctx.Err()
		default:
		}

		// Собираем батч сообщений
		msgs := make([]*sarama.ProducerMessage, 0, len(evs))
		for _, e := range evs {
			topic, key := h.resolve(e)
			msg := &sarama.ProducerMessage{
				Topic:     topic,
				Key:       sarama.StringEncoder(key), // партиционирование по ключу
				Value:     sarama.ByteEncoder(e.Payload),
				Timestamp: e.CreatedAt, // полезно для таймлайнов
				Metadata:  e.ID,        // чтобы распознать ошибку по id
				Headers: []sarama.RecordHeader{
					{
						Key:   []byte(headerEventID),
						Value: []byte(e.ID.String()),
					},
				},
			}
			msgs = append(msgs, msg)
		}

		// Отправляем батчом
		if sendErr := h.producer.SendMessages(msgs); sendErr != nil {
			// Частичные ошибки приходят как sarama.ProducerErrors
			if perrs, ok := sendErr.(sarama.ProducerErrors); ok {
				failedSet := make(map[uuid.UUID]struct{}, len(perrs))
				for _, pe := range perrs {
					log.Println("Write to kafka failed:", pe)

					if pe == nil || pe.Msg == nil {
						continue
					}
					if id, ok2 := pe.Msg.Metadata.(uuid.UUID); ok2 {
						failedSet[id] = struct{}{}
					}
				}
				// Разносим успех/провал по спискам
				for _, m := range msgs {
					id := m.Metadata.(uuid.UUID)
					if _, bad := failedSet[id]; bad {
						failed = append(failed, id)
					} else {
						succeeded = append(succeeded, id)
					}
				}
				// продолжаем другие чанки; общий err вернём последним не-nil
				err = sendErr
				continue
			}

			// Фатальная ошибка всего чанка — считаем всё failed
			for _, m := range msgs {
				failed = append(failed, m.Metadata.(uuid.UUID))
			}
			// Возвращаем ошибку, но часть до этого могла быть отправлена успешно
			return succeeded, failed, sendErr
		}

		// Весь чанк успешен
		for _, m := range msgs {
			succeeded = append(succeeded, m.Metadata.(uuid.UUID))
		}
	}

	return succeeded, failed, err
}

// ===== Helpers =====

func chunkEvents(src []*outbox.Event, size int) [][]*outbox.Event {
	return slices.Collect(slices.Chunk(src, size))
}

func ids(src []*outbox.Event) []uuid.UUID {
	ids := make([]uuid.UUID, len(src))
	for i, e := range src {
		ids[i] = e.ID
	}
	return ids
}
