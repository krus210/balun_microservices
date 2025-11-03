package repository

import (
	"context"
	"fmt"
	"lib/postgres"

	"notifications/internal/app/models"

	"github.com/Masterminds/squirrel"
)

// GetPendingMessagesForProcessing возвращает сообщения, готовые к обработке
// Использует FOR UPDATE SKIP LOCKED для безопасной конкурентной обработки
func (r *Repository) GetPendingMessagesForProcessing(
	ctx context.Context,
	maxAttempts int,
	batchSize int,
) ([]models.InboxMessage, error) {
	const api = "inbox.Repository.GetPendingMessagesForProcessing"

	qb := r.qb.Select(InboxMessagesTableColumns...).
		From(InboxMessagesTable).
		Where(squirrel.Eq{
			InboxMessagesTableColumnStatus: []string{InboxMessageStatusReceived, InboxMessageStatusProcessing, InboxMessageStatusFailed},
		}).
		Where(squirrel.Lt{InboxMessagesTableColumnAttempts: maxAttempts}).
		Suffix("FOR UPDATE SKIP LOCKED").
		Limit(uint64(batchSize))

	conn := r.db.GetQueryEngine(ctx)
	var dbMessages []inboxMessage
	if err := conn.Selectx(ctx, &dbMessages, qb); err != nil {
		return nil, fmt.Errorf("%s: %w", api, postgres.ConvertPGError(err))
	}

	messages := make([]models.InboxMessage, 0, len(dbMessages))
	for i := range dbMessages {
		if msg := ToModel(&dbMessages[i]); msg != nil {
			messages = append(messages, *msg)
		}
	}

	return messages, nil
}
