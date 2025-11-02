package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"lib/postgres"
)

const (
	defaultDeleteTickInterval    = 6 * time.Hour
	defaultDeleteBatchSize       = 1000
	defaultDeleteRetentionPeriod = 24 * time.Hour
)

type deleteRepository interface {
	DeleteOldProcessedMessages(
		ctx context.Context,
		retentionPeriod time.Duration,
		batchSize int,
	) (int64, error)
}

// DeleteWorker воркер для удаления старых обработанных событий из inbox
type DeleteWorker struct {
	repo            deleteRepository
	tm              postgres.TransactionManagerAPI
	tickInterval    time.Duration
	batchSize       int
	retentionPeriod time.Duration
}

// NewDeleteWorker создает новый воркер удаления с настройками по умолчанию
func NewDeleteWorker(repo deleteRepository, tm postgres.TransactionManagerAPI) *DeleteWorker {
	return &DeleteWorker{
		repo:            repo,
		tm:              tm,
		tickInterval:    defaultDeleteTickInterval,
		batchSize:       defaultDeleteBatchSize,
		retentionPeriod: defaultDeleteRetentionPeriod,
	}
}

// WithTickInterval устанавливает интервал опроса
func (w *DeleteWorker) WithTickInterval(interval time.Duration) *DeleteWorker {
	w.tickInterval = interval
	return w
}

// WithBatchSize устанавливает размер батча для удаления
func (w *DeleteWorker) WithBatchSize(size int) *DeleteWorker {
	w.batchSize = size
	return w
}

// WithRetentionPeriod устанавливает период хранения обработанных сообщений
func (w *DeleteWorker) WithRetentionPeriod(period time.Duration) *DeleteWorker {
	w.retentionPeriod = period
	return w
}

// Start запускает воркер удаления с graceful shutdown
func (w *DeleteWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.tickInterval)
	defer ticker.Stop()

	log.Printf("Delete worker started with tick interval: %v, batch size: %d, retention period: %v",
		w.tickInterval, w.batchSize, w.retentionPeriod)

	// Удаляем сразу при старте
	w.deleteOldMessages(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Delete worker stopped gracefully")
			return
		case <-ticker.C:
			w.deleteOldMessages(ctx)
		}
	}
}

// deleteOldMessages удаляет старые обработанные сообщения батчами в цикле
func (w *DeleteWorker) deleteOldMessages(ctx context.Context) {
	log.Printf("Starting deletion of processed messages older than %v", w.retentionPeriod)

	totalDeleted := int64(0)
	iteration := 0

	for {
		iteration++

		var deletedCount int64
		err := w.tm.RunReadCommitted(ctx, func(txCtx context.Context) error {
			count, err := w.repo.DeleteOldProcessedMessages(txCtx, w.retentionPeriod, w.batchSize)
			if err != nil {
				return fmt.Errorf("failed to delete old messages: %w", err)
			}
			deletedCount = count
			return nil
		})
		if err != nil {
			log.Printf("Error deleting old messages (iteration %d): %v", iteration, err)
			break
		}

		if deletedCount == 0 {
			// Больше нечего удалять
			break
		}

		totalDeleted += deletedCount
		log.Printf("Deleted %d messages in iteration %d", deletedCount, iteration)

		// Если удалили меньше чем размер батча, значит больше старых сообщений нет
		if deletedCount < int64(w.batchSize) {
			break
		}

		// Небольшая пауза между батчами чтобы не создавать слишком большую нагрузку на БД
		select {
		case <-ctx.Done():
			log.Printf("Delete worker stopped during batch processing. Total deleted: %d", totalDeleted)
			return
		case <-time.After(100 * time.Millisecond):
			// Продолжаем
		}
	}

	if totalDeleted > 0 {
		log.Printf("Completed deletion cycle: total %d messages deleted in %d iterations", totalDeleted, iteration)
	} else {
		log.Printf("No old messages to delete")
	}
}
