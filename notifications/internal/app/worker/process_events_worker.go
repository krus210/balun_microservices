package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"lib/postgres"

	"notifications/internal/app/models"
	"notifications/internal/app/repository"
)

const (
	defaultTickInterval = 10 * time.Second
	defaultBatchSize    = 100
	defaultMaxAttempts  = 3
)

type inboxRepository interface {
	UpdateInboxMessage(ctx context.Context, params repository.UpdateInboxMessageParams) error
	GetPendingMessagesForProcessing(
		ctx context.Context,
		maxAttempts int,
		batchSize int,
	) ([]models.InboxMessage, error)
}

// SaveEventsWorker воркер для обработки событий из inbox
type SaveEventsWorker struct {
	repo         inboxRepository
	tm           postgres.TransactionManagerAPI
	tickInterval time.Duration
	batchSize    int
	maxAttempts  int
}

// NewSaveEventsWorker создает новый воркер с настройками по умолчанию
func NewSaveEventsWorker(repo inboxRepository, tm postgres.TransactionManagerAPI) *SaveEventsWorker {
	return &SaveEventsWorker{
		repo:         repo,
		tm:           tm,
		tickInterval: defaultTickInterval,
		batchSize:    defaultBatchSize,
		maxAttempts:  defaultMaxAttempts,
	}
}

// WithTickInterval устанавливает интервал опроса
func (w *SaveEventsWorker) WithTickInterval(interval time.Duration) *SaveEventsWorker {
	w.tickInterval = interval
	return w
}

// WithBatchSize устанавливает размер батча
func (w *SaveEventsWorker) WithBatchSize(size int) *SaveEventsWorker {
	w.batchSize = size
	return w
}

// WithMaxAttempts устанавливает максимальное количество попыток
func (w *SaveEventsWorker) WithMaxAttempts(attempts int) *SaveEventsWorker {
	w.maxAttempts = attempts
	return w
}

// Start запускает воркер с graceful shutdown
func (w *SaveEventsWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.tickInterval)
	defer ticker.Stop()

	log.Printf("SaveEventsWorker started with tick interval: %v, batch size: %d, max attempts: %d",
		w.tickInterval, w.batchSize, w.maxAttempts)

	// Обрабатываем сразу при старте
	w.processMessages(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("SaveEventsWorker stopped gracefully")
			return
		case <-ticker.C:
			w.processMessages(ctx)
		}
	}
}

// processMessages обрабатывает батч сообщений в транзакции
func (w *SaveEventsWorker) processMessages(ctx context.Context) {
	err := w.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
		// Получаем сообщения для обработки
		messages, err := w.repo.GetPendingMessagesForProcessing(txCtx, w.maxAttempts, w.batchSize)
		if err != nil {
			return fmt.Errorf("failed to get pending messages: %w", err)
		}

		if len(messages) == 0 {
			return nil
		}

		log.Printf("Processing %d messages", len(messages))

		// Обрабатываем каждое сообщение
		for i := range messages {
			if err := w.processMessage(txCtx, messages[i]); err != nil {
				log.Printf("Failed to process message %s: %v", messages[i].ID, err)
				// Продолжаем обработку остальных сообщений
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("Error processing messages batch: %v", err)
	}
}

// processMessage обрабатывает одно сообщение с защитой от паники
func (w *SaveEventsWorker) processMessage(ctx context.Context, msg models.InboxMessage) (err error) {
	// Защита от паники: откатываем статус в failed при любой ошибке
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
			w.markAsFailed(ctx, msg, err)
		} else if err != nil {
			w.markAsFailed(ctx, msg, err)
		}
	}()

	// Обновляем статус на processing и увеличиваем счетчик попыток
	err = w.repo.UpdateInboxMessage(ctx, repository.UpdateInboxMessageParams{
		ID:       msg.ID,
		Status:   models.InboxMessageStatusProcessing,
		Attempts: msg.Attempts + 1,
	})
	if err != nil {
		return fmt.Errorf("failed to update message to processing: %w", err)
	}

	// Логируем информацию о событии
	log.Printf("Processing event - ID: %s, Topic: %s, Partition: %d, Offset: %d, Payload: %s",
		msg.ID, msg.Topic, msg.Partition, msg.Offset, string(msg.Payload))

	// Здесь можно добавить реальную бизнес-логику обработки события
	// Пока просто эмулируем успешную обработку

	// Обновляем статус на processed с временем обработки
	now := time.Now()
	err = w.repo.UpdateInboxMessage(ctx, repository.UpdateInboxMessageParams{
		ID:          msg.ID,
		Status:      models.InboxMessageStatusProcessed,
		Attempts:    msg.Attempts + 1,
		ProcessedAt: &now,
	})
	if err != nil {
		return fmt.Errorf("failed to update message to processed: %w", err)
	}

	log.Printf("Successfully processed message %s", msg.ID)
	return nil
}

// markAsFailed помечает сообщение как failed
func (w *SaveEventsWorker) markAsFailed(ctx context.Context, msg models.InboxMessage, processingErr error) {
	errMsg := processingErr.Error()
	err := w.repo.UpdateInboxMessage(ctx, repository.UpdateInboxMessageParams{
		ID:        msg.ID,
		Status:    models.InboxMessageStatusFailed,
		Attempts:  msg.Attempts + 1,
		LastError: &errMsg,
	})
	if err != nil {
		log.Printf("Failed to mark message %s as failed: %v", msg.ID, err)
	} else {
		log.Printf("Marked message %s as failed: %v", msg.ID, processingErr)
	}
}
