package repository

import (
	"context"
	"fmt"
	"time"

	"notifications/pkg/postgres"

	"github.com/Masterminds/squirrel"
)

// DeleteOldProcessedMessages удаляет сообщения со статусом processed,
// которые были обработаны более чем retentionPeriod назад.
// Удаление выполняется порциями размером batchSize для предотвращения блокировок БД
func (r *Repository) DeleteOldProcessedMessages(
	ctx context.Context,
	retentionPeriod time.Duration,
	batchSize int,
) (int64, error) {
	const api = "inbox.Repository.DeleteOldProcessedMessages"

	// Вычисляем пороговое время
	thresholdTime := time.Now().Add(-retentionPeriod)

	// Используем подзапрос для выбора ID записей на удаление с LIMIT
	subQuery := r.qb.Select(InboxMessagesTableColumnID).
		From(InboxMessagesTable).
		Where(squirrel.Eq{InboxMessagesTableColumnStatus: InboxMessageStatusProcessed}).
		Where(squirrel.Lt{InboxMessagesTableColumnProcessedAt: thresholdTime}).
		Limit(uint64(batchSize))

	qb := r.qb.Delete(InboxMessagesTable).
		Where(squirrel.Expr(InboxMessagesTableColumnID+" IN (?)", subQuery))

	conn := r.db.GetQueryEngine(ctx)
	commandTag, err := conn.Execx(ctx, qb)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	return commandTag.RowsAffected(), nil
}
